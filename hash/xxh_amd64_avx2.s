//go:build amd64

#include "textflag.h"

// 32-bit prime broadcast table (8 dwords = 32 bytes)
DATA ·prime32+0(SB)/4,  $0x9E3779B1
DATA ·prime32+4(SB)/4,  $0x9E3779B1
DATA ·prime32+8(SB)/4,  $0x9E3779B1
DATA ·prime32+12(SB)/4, $0x9E3779B1
DATA ·prime32+16(SB)/4, $0x9E3779B1
DATA ·prime32+20(SB)/4, $0x9E3779B1
DATA ·prime32+24(SB)/4, $0x9E3779B1
DATA ·prime32+28(SB)/4, $0x9E3779B1
GLOBL ·prime32(SB), RODATA, $32

// void xxh3Accumulate512AVX2(uint64* acc, uint8* in, uint8* secret)
TEXT ·xxh3Accumulate512AVX2(SB), NOSPLIT, $0-24
    // Args
    MOVQ acc+0(FP),    R8     // acc base (8*8 bytes)
    MOVQ in+8(FP),     R9     // input base (64 bytes)
    MOVQ secret+16(FP),R10    // secret base (64 bytes)

    // ---- Block 0: first 32 bytes ----
    // data_vec = *(in + 0)
    VMOVDQU   (R9),      Y0
    // key_vec  = *(secret + 0)
    VMOVDQU   (R10),     Y1
    // data_key = data ^ key
    VPXOR     Y1, Y0,    Y2
    // data_key_lo = data_key >> 32
    VPSRLQ    $32, Y2,   Y3
    // product = (lo32(data_key) * lo32(data_key_lo)) per lane
    VPMULUDQ  Y3, Y2,    Y4
    // data_swap = swap 32-bit halves in each 64-bit lane: _mm256_shuffle_epi32(..., 0x4E)
    VPSHUFD   $0x4E, Y0, Y5
    // sum = acc + data_swap
    VMOVDQU   (R8),      Y6
    VPADDQ    Y5, Y6,    Y6
    // acc = sum + product
    VPADDQ    Y4, Y6,    Y6
    VMOVDQU   Y6, (R8)

    // ---- Block 1: next 32 bytes ----
    VMOVDQU   32(R9),    Y0
    VMOVDQU   32(R10),   Y1
    VPXOR     Y1, Y0,    Y2
    VPSRLQ    $32, Y2,   Y3
    VPMULUDQ  Y3, Y2,    Y4
    VPSHUFD   $0x4E, Y0, Y5
    VMOVDQU   32(R8),    Y6
    VPADDQ    Y5, Y6,    Y6
    VPADDQ    Y4, Y6,    Y6
    VMOVDQU   Y6, 32(R8)

    VZEROUPPER
    RET

// void xxh3ScrambleAccAVX2(uint64* acc, uint8* secret)
TEXT ·xxh3ScrambleAccAVX2(SB), NOSPLIT, $0-16
    MOVQ acc+0(FP),     R8
    MOVQ secret+8(FP),  R10

    // Load prime vector: [p p p p p p p p] as 8 dwords
    VMOVDQU   ·prime32(SB), Y15

    // ---- Block 0: first 32 bytes ----
    // acc_vec
    VMOVDQU   (R8),     Y0
    // shifted = acc >> 47
    VPSRLQ    $47, Y0,  Y1
    // data_vec = acc ^ shifted
    VPXOR     Y1, Y0,   Y2
    // data_key = data_vec ^ secret
    VMOVDQU   (R10),    Y3
    VPXOR     Y3, Y2,   Y4
    // data_key_hi = data_key >> 32
    VPSRLQ    $32, Y4,  Y5
    // prod_lo = mul_epu32(data_key, prime)
    VPMULUDQ  Y15, Y4,  Y6
    // prod_hi = mul_epu32(data_key_hi, prime)
    VPMULUDQ  Y15, Y5,  Y7
    // acc = prod_lo + (prod_hi << 32)
    VPSLLQ    $32, Y7,  Y7
    VPADDQ    Y6, Y7,   Y6
    VMOVDQU   Y6, (R8)

    // ---- Block 1: next 32 bytes ----
    VMOVDQU   32(R8),   Y0
    VPSRLQ    $47, Y0,  Y1
    VPXOR     Y1, Y0,   Y2
    VMOVDQU   32(R10),  Y3
    VPXOR     Y3, Y2,   Y4
    VPSRLQ    $32, Y4,  Y5
    VPMULUDQ  Y15, Y4,  Y6
    VPMULUDQ  Y15, Y5,  Y7
    VPSLLQ    $32, Y7,  Y7
    VPADDQ    Y6, Y7,   Y6
    VMOVDQU   Y6, 32(R8)

    VZEROUPPER
    RET
