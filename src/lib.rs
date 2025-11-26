use std::panic;
use std::ptr;
use std::slice;
use std::os::raw::c_int;

#[no_mangle]
pub extern "C" fn zstd_compress(
    input_ptr: *const u8,
    input_len: usize,
    level: c_int,
    out_ptr: *mut u8,
    out_capacity: usize,
    out_len: *mut usize,
) -> c_int {
    // Return codes:
    // 0 = OK
    // 1 = invalid args
    // 2 = buffer too small (out_len set to required size)
    // 3 = compression error
    // 4 = panic
    if input_ptr.is_null() || input_len == 0 || out_len.is_null() {
        return 1;
    }

    let result = panic::catch_unwind(|| {
        // create input slice
        let input = unsafe { slice::from_raw_parts(input_ptr, input_len) };

        // compress to vec with requested level
        return match zstd::bulk::compress(input, level as i32) {
            Ok(compressed) => {
                let needed = compressed.len();
                unsafe {
                    // write required size back
                    ptr::write(out_len, needed);
                }
                if out_capacity < needed || out_ptr.is_null() || out_capacity == 0 {
                    // caller asked to query size or capacity insufficient
                    return 2;
                }
                // copy into out_ptr
                unsafe {
                    let out_slice = slice::from_raw_parts_mut(out_ptr, out_capacity);
                    out_slice[..needed].copy_from_slice(&compressed);
                    ptr::write(out_len, needed);
                }
                0
            }
            Err(_) => {
                3
            }
        }
    });

    result.unwrap_or_else(|_| 4)
}

#[no_mangle]
pub extern "C" fn zstd_decompress(
    input_ptr: *const u8,
    input_len: usize,
    out_ptr: *mut u8,
    out_capacity: usize,
    out_len: *mut usize,
) -> c_int {
    // Return codes:
    // 0 = OK
    // 1 = invalid args
    // 2 = buffer too small (out_len set to required size)
    // 3 = decompression error
    // 4 = panic
    if input_ptr.is_null() || input_len == 0 || out_len.is_null() {
        return 1;
    }

    let result = panic::catch_unwind(|| {
        let input = unsafe { slice::from_raw_parts(input_ptr, input_len) };

        // 使用一个合理的初始容量，或者使用输入长度作为启发值
        let estimated_capacity = input_len * 3; // 启发式估计，可根据需要调整

        return match zstd::bulk::decompress(input, estimated_capacity) {
            Ok(decompressed) => {
                let needed = decompressed.len();
                unsafe {
                    ptr::write(out_len, needed);
                }
                if out_capacity < needed || out_ptr.is_null() || out_capacity == 0 {
                    return 2;
                }
                unsafe {
                    let out_slice = slice::from_raw_parts_mut(out_ptr, out_capacity);
                    out_slice[..needed].copy_from_slice(&decompressed);
                    ptr::write(out_len, needed);
                }
                0
            }
            Err(_) => {
                3
            }
        }
    });

    result.unwrap_or_else(|_| 4)
}
