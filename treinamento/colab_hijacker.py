import torch
import torch.nn as nn
import torch.optim as optim
import hashlib
import struct
from collections import Counter
import re
import os

try:
    from datasets import load_dataset
    from transformers import AutoTokenizer, AutoModel
except:
    os.system("pip install datasets transformers")
    from datasets import load_dataset
    from transformers import AutoTokenizer, AutoModel

# =========================================================================
# FASE 27 - CROM LSH-MoE + NEURAL HIJACKING (EMBEDDING SUBSPACE DISTILLATION)
# =========================================================================

EPOCHS = 150 # Na dimensão 1024, 150 já é massivo
BATCH_SIZE = 16384  # A100 UNCHAINED LEVEL 2: Saturando os 40GB VRAM com superblocos
DIM = 1024         
NUM_EXPERTS = 128  
NUM_LAYERS = 3     # Evolução DeepStack: 3 Camadas O(1) Empilhadas
VOCAB_SIZE = 32000 
SEQ_LEN = 16 
MAX_TEXTS = 200000 

device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
print(f"[*] CROM-Engine Iniciado. Hardware: {device}")

# 1. Dataset de Experiência
print("[*] Baixando Conhecimento Base (Alpaca / Instruct)...")
real_data = load_dataset("tatsu-lab/alpaca", split=f"train[:{MAX_TEXTS}]")

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

VOCAB_SIZE = len(vocab) 

def tokenize(text):
    clean_words = re.findall(r'\b\w+\b|[.!?]', text.lower())
    return [vocab.get(w, 0) for w in clean_words]

def compute_expert_id(context_tokens):
    return sum(context_tokens) % NUM_EXPERTS

print(f"[*] Janela de Roteamento MoE Ajustada: SEQ_LEN = {SEQ_LEN}")
X_experts = {i: [] for i in range(NUM_EXPERTS)}
Y_experts = {i: [] for i in range(NUM_EXPERTS)}

for d in dataset:
    text = f"{d['instruction']} {d['input']} <think> Raciocínio... </think> {d['output']}"
    tokens = tokenize(text)
    padded_tokens = [1] * SEQ_LEN + tokens
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

# Arquitetura Plana O(1) do Crom
class MultiHeadGlobalAttention(nn.Module):
    def __init__(self, dim):
        super().__init__()
        self.W_q = nn.Linear(dim, dim, bias=False)
        self.W_k = nn.Linear(dim, dim, bias=False)
        self.W_v = nn.Linear(dim, dim, bias=False)
        self.dim = dim
        
    def forward(self, x):
        q = self.W_q(x[:, -1, :]).unsqueeze(1) 
        k = self.W_k(x)                        
        v = self.W_v(x)
        scores = torch.bmm(q, k.transpose(1, 2)) / (self.dim ** 0.5)
        attn = torch.softmax(scores, dim=-1)   
        context = torch.bmm(attn, v)           
        return context.squeeze(1)

class CROM_Layer(nn.Module):
    def __init__(self):
        super().__init__()
        self.attention = MultiHeadGlobalAttention(DIM)
        # Intermediate MoE: Mantém o Fuso Dimensional O(1)
        self.experts = nn.ModuleList([nn.Linear(DIM, DIM) for _ in range(NUM_EXPERTS)])
        self.norm = nn.LayerNorm(DIM)
        
    def forward(self, x_emb, expert_id):
        attn_vec = self.attention(x_emb)
        # Residual Clássica para blindar o vanishing gradient multi-layer
        out = self.experts[expert_id](attn_vec)
        # Atualiza APENAS o Token focado (Causal Update)
        new_focus = self.norm(x_emb[:, -1, :] + out)
        x_copy = x_emb.clone()
        x_copy[:, -1, :] = new_focus
        return x_copy

class CROM_MoE(nn.Module):
    def __init__(self):
        super().__init__()
        self.embedding = nn.Embedding(VOCAB_SIZE, DIM)
        self.pos_embedding = nn.Embedding(SEQ_LEN, DIM) # A Cura da Cegueira Sintática!
        self.layers = nn.ModuleList([CROM_Layer() for _ in range(NUM_LAYERS)])
        self.lm_head = nn.Linear(DIM, VOCAB_SIZE) # Decodificador Semântico Final
        
    def forward(self, x, expert_id):
        seq_len_current = x.shape[1]
        positions = torch.arange(seq_len_current, device=device).unsqueeze(0)
        
        word_emb = self.embedding(x)
        pos_emb = self.pos_embedding(positions)
        
        # Fusão Tempo + Espaço
        x_emb = word_emb + pos_emb
        
        # O Passe Profundo (Deep Stacking)
        for layer in self.layers:
            x_emb = layer(x_emb, expert_id)
            
        final_context = x_emb[:, -1, :]
        return self.lm_head(final_context)

model = CROM_MoE().to(device)

# =========================================================================================
# FASE 27 HIJACKING: Extrair Consciência de Redes Gigantes (Transplante Semântico)
# =========================================================================================
print("\n[!!!] INICIANDO DECODIFICAÇÃO E HIJACKING DO MESTRE (QWEN) [!!!]")
print("[*] Estabelecendo conexão com HuggingFace para Sugar Vetores Semânticos do Qwen 1.5B...")
TEACHER_ID = "Qwen/Qwen2.5-0.5B"
try:
    hf_tokenizer = AutoTokenizer.from_pretrained(TEACHER_ID)
    hf_model = AutoModel.from_pretrained(TEACHER_ID)
    hf_model = hf_model.to(device) # Coloca na GPU pra extrair mais rapido
    qwen_embeddings = hf_model.get_input_embeddings()
    QWEN_DIM = hf_model.config.hidden_size # Provavelmente 896
    print(f"[*] Transplante pronto! Dimensão Qwen: {QWEN_DIM}. Projetando matematicamente para CROM DIM: {DIM}...")
    
    # Adotando a DType exata do modelo mestre para não colidir tensores (Ex: BFloat16)
    dtype_mestre = qwen_embeddings.weight.dtype
    funnel_projection = torch.randn(QWEN_DIM, DIM, device=device, dtype=dtype_mestre) / (DIM**0.5)
    
    with torch.no_grad():
        hijack_count = 0
        for word, crom_id in vocab.items():
            # O Segredo: Deixamos o Qwen interpretar e subdividir nossas strings do vocabulario Go.
            qwen_ids = hf_tokenizer.encode(word, add_special_tokens=False)
            if len(qwen_ids) == 0: continue
            
            qwen_vecs = qwen_embeddings(torch.tensor(qwen_ids, device=device)) # Vetores maciços da nuvem
            mean_vec = qwen_vecs.mean(dim=0) # Contrai em 1 a inteligência de múltiplas sub-words
            
            # Compressão Multi-Dimensional (PCA simplificado LSH) reduzindo pro DIM256 do nosso Motor Pleno
            crom_vec = torch.matmul(mean_vec, funnel_projection)
            
            # O TRANSPLANTE NEUROMÓRFICO! A palavra nunca mais sofrerá de amnésia semântica.
            model.embedding.weight.data[crom_id] = crom_vec.to(dtype=torch.float32) # Volta pro Array do CROM
            hijack_count += 1
            
    print(f"[*] SUCESSO DO HIJACKING! {hijack_count} tokens injetados com a Consciência de Mundo do Qwen!")
    
    # Destravamos a Rede Base! 
    model.embedding.weight.requires_grad = True
    print("[*] Consciência Mestra Extraída. Iniciando Ajuste Fino LSH Ativo (Descongelado)...")
    
    # [SRE FIX] Desalocar o Modelo Qwen da VRAM da GPU para liberar espaço pro Backprop
    del hf_model
    del hf_tokenizer
    torch.cuda.empty_cache()
    print("[*] VRAM da Mestra Desalocada. Espaço libre garantido.")
except Exception as e:
    print(f"[!] Erro no Neural Hijack, operando no modo de Aprendizagem Virgem CROM (Scratch). Erro: {e}")
# =========================================================================================

optimizer = optim.AdamW(filter(lambda p: p.requires_grad, model.parameters()), lr=0.0003) 
loss_fn = nn.CrossEntropyLoss()
scaler = torch.amp.GradScaler('cuda') # Habilita cálculo adaptativo BFloat16

print("\n[*] Iniciando Queima Neural Acelerada MAX-TURBO (Tensor Cores)...")
for epoch in range(EPOCHS):
    total_loss = 0
    t_count = 0
    for exp_id in range(NUM_EXPERTS):
        X = X_experts[exp_id]
        Y = Y_experts[exp_id]
        if type(X) == list or len(X) == 0: continue
        
        # [SRE FIX MAX-TURBO] Batch massivo + Tensor Cores A100 (bfloat16)
        exper_loss = 0
        b_count = 0
        for i in range(0, len(X), BATCH_SIZE):
            batch_X = X[i:i+BATCH_SIZE]
            batch_Y = Y[i:i+BATCH_SIZE]
            
            optimizer.zero_grad()
            
            # Autocast faz o PyTorch utilizar os núcleos neurais de silício puro da A100
            with torch.amp.autocast('cuda', dtype=torch.bfloat16):
                logits = model(batch_X, exp_id)
                loss = loss_fn(logits, batch_Y)
            
            scaler.scale(loss).backward()
            scaler.step(optimizer)
            scaler.update()
            
            exper_loss += loss.item()
            b_count += 1
            
        total_loss += exper_loss / b_count
        t_count += 1
        
    print(f"    Epoch {epoch+1}/{EPOCHS} -> Attention Loss Média Combinada: {total_loss/t_count:.4f}")

print("\n[*] Quantizando Tensor Inteligente e Criando Formato Crom Int32/Int8 (BitNet Lógica)...")
with torch.no_grad():
    max_val = model.embedding.weight.abs().max()
    max_val = max(max_val, model.pos_embedding.weight.abs().max())
    max_val = max(max_val, model.lm_head.weight.abs().max())
    for layer in model.layers:
        max_val = max(max_val, layer.attention.W_q.weight.abs().max())
        max_val = max(max_val, layer.attention.W_k.weight.abs().max())
        max_val = max(max_val, layer.attention.W_v.weight.abs().max())
        for exp in layer.experts:
            max_val = max(max_val, exp.weight.abs().max())
        
    scale = (max_val / 127.0).item()
    if scale == 0: scale = 0.0001
    inv_scale = 1.0 / scale
    
    q_emb = torch.clamp(torch.round(model.embedding.weight * inv_scale), -127, 127).to(torch.int8)
    q_pos = torch.clamp(torch.round(model.pos_embedding.weight * inv_scale), -127, 127).to(torch.int8)
    q_lm_head = torch.clamp(torch.round(model.lm_head.weight * inv_scale), -127, 127).to(torch.int8)
    q_lm_bias = torch.round(model.lm_head.bias / (scale * scale)).to(torch.int32)

    with open("brain.crom", "wb") as f:
        f.write(struct.pack("<I", VOCAB_SIZE))
        f.write(struct.pack("<I", SEQ_LEN)) 
        
        for w in rev_vocab:
            b_word = w.encode('utf-8')
            f.write(struct.pack("<H", len(b_word)))
            f.write(b_word)
            
        f.write(struct.pack("<I", DIM))
        f.write(struct.pack("<I", NUM_EXPERTS))
        f.write(struct.pack("<I", NUM_LAYERS)) # NEW: Guardiões da Profundidade
        f.write(struct.pack("<f", scale))
        
        f.write(q_emb.cpu().numpy().tobytes())
        f.write(q_pos.cpu().numpy().tobytes()) 
        
        for layer in model.layers:
            q_wq = torch.clamp(torch.round(layer.attention.W_q.weight * inv_scale), -127, 127).to(torch.int8)
            q_wk = torch.clamp(torch.round(layer.attention.W_k.weight * inv_scale), -127, 127).to(torch.int8)
            q_wv = torch.clamp(torch.round(layer.attention.W_v.weight * inv_scale), -127, 127).to(torch.int8)
            f.write(q_wq.cpu().numpy().tobytes())
            f.write(q_wk.cpu().numpy().tobytes())
            f.write(q_wv.cpu().numpy().tobytes())
            
            # Fase 29: Exportação bruta dos Pesos de LayerNorm em Float32 para precisão
            f.write(layer.norm.weight.cpu().float().numpy().tobytes())
            f.write(layer.norm.bias.cpu().float().numpy().tobytes())
            
            for exp in layer.experts:
                q_w = torch.clamp(torch.round(exp.weight * inv_scale), -127, 127).to(torch.int8)
                q_b = torch.round(exp.bias / (scale * scale)).to(torch.int32)
                f.write(q_w.cpu().numpy().tobytes())
                f.write(q_b.cpu().numpy().tobytes())
                
        f.write(q_lm_head.cpu().numpy().tobytes())
        f.write(q_lm_bias.cpu().numpy().tobytes())
print("[*] Transplante Neurológico CROM-LLM Concluído e exportado (brain.crom)!")

# =========================================================================
# FASE BÔNUS: COLAB BACKUP
# =========================================================================
print("\n[*] Salvando Cérebro no Google Drive para segurança...")
try:
    from google.colab import drive
    drive.mount('/content/drive')
    
    # Cria a pasta caso não exista
    os.makedirs("/content/drive/MyDrive/CROM_Models", exist_ok=True)
    
    # Copia o arquivo crom de modo instantâneo
    os.system("cp brain.crom '/content/drive/MyDrive/CROM_Models/brain_hijacked.crom'")
    print("[*] BACKUP SEGURO E CONCLUÍDO! O Modelo está na sua pasta 'CROM_Models' do Google Drive.")
except Exception as e:
    print(f"[!] Aviso SRE: O ambiente não parece ser o Google Colab ou a permissão do Drive foi negada. Erro: {e}")
