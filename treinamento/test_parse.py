import struct
with open("brain.crom", "rb") as f:
    vocab_size = struct.unpack("<I", f.read(4))[0]
    seq_len = struct.unpack("<I", f.read(4))[0]
    print(f"Vocab: {vocab_size}, SeqLen: {seq_len}")
    
    for _ in range(vocab_size):
        w_len = struct.unpack("<H", f.read(2))[0]
        f.read(w_len)
        
    dim = struct.unpack("<I", f.read(4))[0]
    num_exp = struct.unpack("<I", f.read(4))[0]
    num_lay = struct.unpack("<I", f.read(4))[0]
    scale = struct.unpack("<f", f.read(4))[0]
    print(f"Dim: {dim}, NumExperts: {num_exp}, NumLayers: {num_lay}, Scale: {scale}")
