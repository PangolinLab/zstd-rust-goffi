#ifndef ZSTD_INTERFACE_H
#define ZSTD_INTERFACE_H

#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

// Compress input_ptr[0..input_len) with given level.
// If out_ptr == NULL or out_capacity == 0, function will set *out_len to required size and return 2.
// Return codes:
// 0 = OK
// 1 = invalid args
// 2 = buffer too small (or query) -> *out_len = required size
// 3 = compression error
// 4 = panic
int zstd_compress(
    const unsigned char* input_ptr,
    size_t input_len,
    int level,
    unsigned char* out_ptr,
    size_t out_capacity,
    size_t* out_len
);

// Decompress input_ptr[0..input_len) similarly.
// If out_ptr == NULL or out_capacity == 0, function will set *out_len to required size and return 2.
// Return codes:
// 0 = OK
// 1 = invalid args
// 2 = buffer too small (or query) -> *out_len = required size
// 3 = decompression error
// 4 = panic
int zstd_decompress(
    const unsigned char* input_ptr,
    size_t input_len,
    unsigned char* out_ptr,
    size_t out_capacity,
    size_t* out_len
);

#ifdef __cplusplus
}
#endif

#endif // ZSTD_INTERFACE_H
