/*-
 * SPDX-License-Identifier: MIT
 * SPDX-FileCopyrightText: Copyright (c) 2022 Frank Smit <https://61924.nl/>
 */

#ifndef BARE_H
#define BARE_H

#include <stdint.h>
#include <stdbool.h>

typedef enum {
	BARE_ERROR_NONE,
	BARE_ERROR_WRITE_FAILED,
	BARE_ERROR_READ_FAILED,
	BARE_ERROR_BUFFER_TOO_SMALL,
	BARE_ERROR_INVALID_UTF8,
} bare_error;

typedef bare_error (*bare_write_func)(void *buffer, void *src, uint64_t sz);
typedef bare_error (*bare_read_func)(void *buffer, void *dst, uint64_t sz);

struct bare_writer {
	void *buffer;
	bare_write_func write;
};

struct bare_reader {
	void *buffer;
	bare_read_func read;
};

bare_error bare_put_uint(struct bare_writer *ctx, uint64_t x); /* varuint */
bare_error bare_get_uint(struct bare_reader *ctx, uint64_t *x); /* varuint */
bare_error bare_put_u8(struct bare_writer *ctx, uint8_t x);
bare_error bare_get_u8(struct bare_reader *ctx, uint8_t *x);
bare_error bare_put_u16(struct bare_writer *ctx, uint16_t x);
bare_error bare_get_u16(struct bare_reader *ctx, uint16_t *x);
bare_error bare_put_u32(struct bare_writer *ctx, uint32_t x);
bare_error bare_get_u32(struct bare_reader *ctx, uint32_t *x);
bare_error bare_put_u64(struct bare_writer *ctx, uint64_t x);
bare_error bare_get_u64(struct bare_reader *ctx, uint64_t *x);

bare_error bare_put_int(struct bare_writer *ctx, int64_t x); /* varint */
bare_error bare_get_int(struct bare_reader *ctx, int64_t *x); /* varint */
bare_error bare_put_i8(struct bare_writer *ctx, int8_t x);
bare_error bare_get_i8(struct bare_reader *ctx, int8_t *x);
bare_error bare_put_i16(struct bare_writer *ctx, int16_t x);
bare_error bare_get_i16(struct bare_reader *ctx, int16_t *x);
bare_error bare_put_i32(struct bare_writer *ctx, int32_t x);
bare_error bare_get_i32(struct bare_reader *ctx, int32_t *x);
bare_error bare_put_i64(struct bare_writer *ctx, int64_t x);
bare_error bare_get_i64(struct bare_reader *ctx, int64_t *x);

bare_error bare_put_f32(struct bare_writer *ctx, float x);
bare_error bare_get_f32(struct bare_reader *ctx, float *x);
bare_error bare_put_f64(struct bare_writer *ctx, double x);
bare_error bare_get_f64(struct bare_reader *ctx, double *x);

bare_error bare_put_bool(struct bare_writer *ctx, bool x);
bare_error bare_get_bool(struct bare_reader *ctx, bool *x);

bare_error bare_put_fixed_data(struct bare_writer *ctx, uint8_t *src, uint64_t sz);
bare_error bare_get_fixed_data(struct bare_reader *ctx, uint8_t *dst, uint64_t sz);
bare_error bare_put_data(struct bare_writer *ctx, uint8_t *src, uint64_t sz);
bare_error bare_get_data(struct bare_reader *ctx, uint8_t *dst, uint64_t sz);
bare_error bare_put_str(struct bare_writer *ctx, char *src, uint64_t sz);
bare_error bare_get_str(struct bare_reader *ctx, char *dst, uint64_t sz);

#endif /* BARE_H */
