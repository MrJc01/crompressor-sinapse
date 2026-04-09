import json
import torch
import torch.nn as nn
import torch.optim as optim
import hashlib
import struct
from collections import Counter
import re

# =========================================================================
# CROM LSH-MoE (Sparse Mixture of Experts) + ATENÇÃO!
# =========================================================================

EPOCHS = 350 # Aumentado para descer a Loss com precisão
BATCH_SIZE = 32
DIM = 256          
NUM_EXPERTS = 64  
VOCAB_SIZE = 8192 
SEQ_LEN = 16 # O Foco Cognitivo Agora Avalia 16 Palavras de Desdobramento Semântico
MAX_TEXTS = 300 

device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
print(f"[*] Motores Ativos (Treinador): {device}")

# O DATASET REAL (SEM BRINCADEIRA)
try:
    from datasets import load_dataset
except:
    import os
    os.system("pip install datasets")
    from datasets import load_dataset

print("[*] Baixando Conhecimento de Mundo Real (Alpaca Português)...")
# Puxando as primeiras 500 interações avançadas de um dataset instrucional Real
real_data = load_dataset("tatsu-lab/alpaca", split="train[:300]")

dataset = []
for row in real_data:
    dataset.append({
        "instruction": row["instruction"],
        "input": row["input"],
        "output": row["output"]
    })

all_words = []
for d in dataset:
    text = f"{d['instruction']} {d['input']} <think> Raciocínio... </think> {d['output']}".lower()
    clean_words = re.findall(r'\b\w+\b|[.!?]', text)
    all_words.extend(clean_words)

counts = Counter(all_words)
most_common = counts.most_common(VOCAB_SIZE - 2)
vocab = {"<UNK>": 0, "<PAD>": 1}
rev_vocab = ["<UNK>", "<PAD>"]

for word, count in most_common:
    vocab[word] = len(vocab)
    rev_vocab.append(word)

VOCAB_SIZE = len(vocab) # FIX CRÍTICO: Reduz as matrizes ao tamanho exato dos tokens achados


def tokenize(text):
    clean_words = re.findall(r'\b\w+\b|[.!?]', text.lower())
    return [vocab.get(w, 0) for w in clean_words]

def compute_expert_id(context_tokens):
    return sum(context_tokens) % NUM_EXPERTS

print("[*] Fragmentando Conhecimento (Sliding Windows)...")
X_experts = {i: [] for i in range(NUM_EXPERTS)}
Y_experts = {i: [] for i in range(NUM_EXPERTS)}

for d in dataset:
    text = f"{d['instruction']} {d['input']} <think> Raciocínio... </think> {d['output']}"
    tokens = tokenize(text)
    
    # Preenche com PAD para garantir que as primeiras palavras da frase também tenham o mesmo shape
    padded_tokens = [1] * SEQ_LEN + tokens
    
    # Sliding Window Dinâmico (Sempre olhando para as últimas SEQ_LEN palavras puras)
    for i in range(SEQ_LEN, len(padded_tokens)):
        ctx = padded_tokens[i-SEQ_LEN:i]
        target = padded_tokens[i]
        exp_id = compute_expert_id(ctx)
        X_experts[exp_id].append(ctx)
        Y_experts[exp_id].append(target)

for i in range(NUM_EXPERTS):
    if len(X_experts[i]) > 0:
        X_experts[i] = torch.tensor(X_experts[i], dtype=torch.long, device=device)
        Y_experts[i] = torch.tensor(Y_experts[i], dtype=torch.long, device=device)

# A GRANDE ADIÇÃO: O Self-Attention Global Customizado
class MultiHeadGlobalAttention(nn.Module):
    def __init__(self, dim):
        super().__init__()
        # As matrizes puras do transformer, pra escalar O(1) no Go depois!
        self.W_q = nn.Linear(dim, dim, bias=False)
        self.W_k = nn.Linear(dim, dim, bias=False)
        self.W_v = nn.Linear(dim, dim, bias=False)
        self.dim = dim
        
    def forward(self, x):
        # x.shape = (Batch, Seq, Dim)
        q = self.W_q(x[:, -1, :]).unsqueeze(1) # Extrai e Foca no ULTIMO token gerado (O Ponto focal da predição Causal)
        k = self.W_k(x)                        # (Batch, Seq, Dim)
        v = self.W_v(x)
        
        # Softmax( Q * K^T / sqrt(d) ) 
        scores = torch.bmm(q, k.transpose(1, 2)) / (self.dim ** 0.5)
        attn = torch.softmax(scores, dim=-1)   # (Batch, 1, Seq)
        
        # C = attn * V
        context = torch.bmm(attn, v)           # (Batch, 1, Dim)
        return context.squeeze(1)

class CROM_MoE(nn.Module):
    def __init__(self):
        super().__init__()
        self.embedding = nn.Embedding(VOCAB_SIZE, DIM)
        self.attention = MultiHeadGlobalAttention(DIM)
        # Os Roteadores Lineares
        self.experts = nn.ModuleList([nn.Linear(DIM, VOCAB_SIZE) for _ in range(NUM_EXPERTS)])
        
    def forward(self, x, expert_id):
        emb = self.embedding(x)
        # Em vez de tirar a média Míope, agora a Inteligência foca nos termos relacionais!
        attn_vec = self.attention(emb)
        return self.experts[expert_id](attn_vec)

model = CROM_MoE().to(device)
optimizer = optim.AdamW(model.parameters(), lr=0.0003) # FIX SRE: LR de 0.003 para 0.0003 para parar de quicar na Loss 1.0!
loss_fn = nn.CrossEntropyLoss()

print("[*] Iniciando Queima Neural MoE-Attention (SGD Cuda)...")
for epoch in range(EPOCHS):
    total_loss = 0
    t_count = 0
    
    for exp_id in range(NUM_EXPERTS):
        X = X_experts[exp_id]
        Y = Y_experts[exp_id]
        if type(X) == list or len(X) == 0: continue
        
        logits = model(X, exp_id)
        loss = loss_fn(logits, Y)
        
        optimizer.zero_grad()
        loss.backward()
        optimizer.step()
        
        total_loss += loss.item()
        t_count += 1
        
    print(f"    Epoch {epoch+1}/{EPOCHS} -> Attention Loss Média Combinada: {total_loss/t_count:.4f}")

# Quantiza e Exporta com os Tensor Weights da Attention acoplados!
print("[*] Quantizando Tensor e Criando Formato Crom Int32/Int8 (BitNet Lógica)...")
with torch.no_grad():
    max_val = model.embedding.weight.abs().max()
    max_val = max(max_val, model.attention.W_q.weight.abs().max())
    max_val = max(max_val, model.attention.W_k.weight.abs().max())
    max_val = max(max_val, model.attention.W_v.weight.abs().max())
    for exp in model.experts:
        max_val = max(max_val, exp.weight.abs().max())
        
    scale = (max_val / 127.0).item()
    if scale == 0: scale = 0.0001
    inv_scale = 1.0 / scale
    
    q_emb = torch.clamp(torch.round(model.embedding.weight * inv_scale), -127, 127).to(torch.int8)
    q_wq = torch.clamp(torch.round(model.attention.W_q.weight * inv_scale), -127, 127).to(torch.int8)
    q_wk = torch.clamp(torch.round(model.attention.W_k.weight * inv_scale), -127, 127).to(torch.int8)
    q_wv = torch.clamp(torch.round(model.attention.W_v.weight * inv_scale), -127, 127).to(torch.int8)

    with open("brain.crom", "wb") as f:
        f.write(struct.pack("<I", VOCAB_SIZE))
        for w in rev_vocab:
            b_word = w.encode('utf-8')
            f.write(struct.pack("<H", len(b_word)))
            f.write(b_word)
            
        f.write(struct.pack("<I", DIM))
        f.write(struct.pack("<I", NUM_EXPERTS))
        f.write(struct.pack("<f", scale))
        
        f.write(q_emb.cpu().numpy().tobytes())
        # Grava Matrizes da Atencao (Para os multiplicadores Int8 no Go executarem Softmax Relacional)
        f.write(q_wq.cpu().numpy().tobytes())
        f.write(q_wk.cpu().numpy().tobytes())
        f.write(q_wv.cpu().numpy().tobytes())
        
        for exp in model.experts:
            q_w = torch.clamp(torch.round(exp.weight * inv_scale), -127, 127).to(torch.int8)
            q_b = torch.round(exp.bias / (scale * scale)).to(torch.int32)
            f.write(q_w.cpu().numpy().tobytes())
            f.write(q_b.cpu().numpy().tobytes())
