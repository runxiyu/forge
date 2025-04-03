/*-
 * SPDX-License-Identifier: MIT
 * SPDX-FileCopyrightText: Copyright (c) 2022 Frank Smit <https://61924.nl/>
 */

#include <string.h>
#include <stdbool.h>

#include "bare.h"
#include "utf8.h"

#define UNUSED(x) (void)(x)

enum {
	U8SZ = 1,
	U16SZ = 2,
	U32SZ = 4,
	U64SZ = 8,
	MAXVARINTSZ = 10,
};

static bool
checkstr(const char *x, uint64_t sz)
{
	if (x == NULL || sz == 0) {
		return true;
	}

	int err = 0;
	uint32_t cp = 0;
	char *buf = (void *)x;
	uint64_t chunk = 4;
	char *pad = (char *)(char[4]){0, 0, 0, 0};

#define _utf8_decode(buf) \
	do { \
		buf = utf8_decode(buf, &cp, &err); \
		if (err > 0) { \
			return false; \
		} \
	} while (0)

	for (; sz >= chunk; sz -= chunk) {
		_utf8_decode(buf);
	}

	if (sz > 0) {
		memcpy(pad, buf, sz);
		_utf8_decode(pad);
	}

#undef _utf8_decode

	return true;
}

bare_error
bare_put_uint(struct bare_writer *ctx, uint64_t x)
{
	uint64_t i = 0;
	uint8_t b[MAXVARINTSZ];

	while (x >= 0x80) {
		b[i] = (uint8_t)x | 0x80;
		x >>= 7;
		i++;
	}

	b[i] = (uint8_t)x;
	i++;

	return ctx->write(ctx->buffer, b, i);
}

bare_error
bare_get_uint(struct bare_reader *ctx, uint64_t *x)
{
	bare_error err = BARE_ERROR_NONE;

	uint8_t shift = 0;
	uint64_t result = 0;

	for (uint8_t i = 0;i < 10;i++) {
		uint8_t b;

		err = ctx->read(ctx->buffer, &b, U8SZ);
		if (err != BARE_ERROR_NONE) {
			break;
		}

		if (b < 0x80) {
			result |= (uint64_t)b << shift;
			break;
		} else {
			result |= ((uint64_t)b & 0x7f) << shift;
			shift += 7;
		}
	}

	*x = result;

	return err;
}

bare_error
bare_put_int(struct bare_writer *ctx, int64_t x)
{
	uint64_t ux = (uint64_t)x << 1;

	if (x < 0) {
		ux = ~ux;
	}

	return bare_put_uint(ctx, ux);
}

bare_error
bare_get_int(struct bare_reader *ctx, int64_t *x)
{
	uint64_t ux;

	bare_error err = bare_get_uint(ctx, &ux);

	if (err == BARE_ERROR_NONE) {
		*x = (int64_t)(ux >> 1);

		if ((ux & 1) != 0) {
			*x = ~(*x);
		}
	}

	return err;
}

bare_error
bare_put_u8(struct bare_writer *ctx, uint8_t x)
{
	return ctx->write(ctx->buffer, &x, U8SZ);
}

bare_error
bare_get_u8(struct bare_reader *ctx, uint8_t *x)
{
	return ctx->read(ctx->buffer, x, U8SZ);
}

bare_error
bare_put_u16(struct bare_writer *ctx, uint16_t x)
{
	return ctx->write(ctx->buffer, (uint8_t[U16SZ]){x, x >> 8}, U16SZ);
}

bare_error
bare_get_u16(struct bare_reader *ctx, uint16_t *x)
{
	bare_error err = ctx->read(ctx->buffer, x, U16SZ);

	if (err == BARE_ERROR_NONE) {
		*x = (uint16_t)((uint8_t *)x)[0]
		   | (uint16_t)((uint8_t *)x)[1] << 8;
	}

	return err;
}

bare_error
bare_put_u32(struct bare_writer *ctx, uint32_t x)
{
	uint8_t buf[U32SZ];

	buf[0] = (uint8_t)(x);
	buf[1] = (uint8_t)(x >> 8);
	buf[2] = (uint8_t)(x >> 16);
	buf[3] = (uint8_t)(x >> 24);

	return ctx->write(ctx->buffer, buf, U32SZ);
}

bare_error
bare_get_u32(struct bare_reader *ctx, uint32_t *x)
{
	bare_error err = ctx->read(ctx->buffer, x, U32SZ);

	if (err == BARE_ERROR_NONE) {
		*x = (uint32_t)(((uint8_t *)x)[0])
		   | (uint32_t)(((uint8_t *)x)[1] << 8)
		   | (uint32_t)(((uint8_t *)x)[2] << 16)
		   | (uint32_t)(((uint8_t *)x)[3] << 24);
	}

	return err;
}

bare_error
bare_put_u64(struct bare_writer *ctx, uint64_t x)
{
	uint8_t buf[U64SZ];

	buf[0] = x;
	buf[1] = x >> 8;
	buf[2] = x >> 16;
	buf[3] = x >> 24;
	buf[4] = x >> 32;
	buf[5] = x >> 40;
	buf[6] = x >> 48;
	buf[7] = x >> 56;

	return ctx->write(ctx->buffer, buf, U64SZ);
}

bare_error
bare_get_u64(struct bare_reader *ctx, uint64_t *x)
{
	bare_error err = ctx->read(ctx->buffer, x, U64SZ);

	if (err == BARE_ERROR_NONE) {
		*x = (uint64_t)((uint8_t *)x)[0]
		   | (uint64_t)((uint8_t *)x)[1] << 8
		   | (uint64_t)((uint8_t *)x)[2] << 16
		   | (uint64_t)((uint8_t *)x)[3] << 24
		   | (uint64_t)((uint8_t *)x)[4] << 32
		   | (uint64_t)((uint8_t *)x)[5] << 40
		   | (uint64_t)((uint8_t *)x)[6] << 48
		   | (uint64_t)((uint8_t *)x)[7] << 56;
	}

	return err;
}

bare_error
bare_put_i8(struct bare_writer *ctx, int8_t x)
{
	return bare_put_u8(ctx, x);
}

bare_error
bare_get_i8(struct bare_reader *ctx, int8_t *x)
{
	return bare_get_u8(ctx, (uint8_t *)x);
}

bare_error
bare_put_i16(struct bare_writer *ctx, int16_t x)
{
	return bare_put_u16(ctx, x);
}

bare_error
bare_get_i16(struct bare_reader *ctx, int16_t *x)
{
	return bare_get_u16(ctx, (uint16_t *)x);
}

bare_error
bare_put_i32(struct bare_writer *ctx, int32_t x)
{
	return bare_put_u32(ctx, x);
}

bare_error
bare_get_i32(struct bare_reader *ctx, int32_t *x)
{
	return bare_get_u32(ctx, (uint32_t *)x);
}

bare_error
bare_put_i64(struct bare_writer *ctx, int64_t x)
{
	return bare_put_u64(ctx, x);
}

bare_error
bare_get_i64(struct bare_reader *ctx, int64_t *x)
{
	return bare_get_u64(ctx, (uint64_t *)x);
}

bare_error
bare_put_f32(struct bare_writer *ctx, float x)
{
	uint32_t b;
	memcpy(&b, &x, U32SZ);

	return bare_put_u32(ctx, b);
}

bare_error
bare_get_f32(struct bare_reader *ctx, float *x)
{
	return ctx->read(ctx->buffer, x, U32SZ);
}

bare_error
bare_put_f64(struct bare_writer *ctx, double x)
{
	uint64_t b;
	memcpy(&b, &x, U64SZ);

	return bare_put_u64(ctx, b);
}

bare_error
bare_get_f64(struct bare_reader *ctx, double *x)
{
	return ctx->read(ctx->buffer, x, U64SZ);
}

bare_error
bare_put_bool(struct bare_writer *ctx, bool x)
{
	return bare_put_u8(ctx, (uint8_t)x);
}

bare_error
bare_get_bool(struct bare_reader *ctx, bool *x)
{
	return bare_get_u8(ctx, (uint8_t *)x);
}

bare_error
bare_put_fixed_data(struct bare_writer *ctx, uint8_t *src, uint64_t sz)
{
	return ctx->write(ctx->buffer, (void *)src, sz);
}

bare_error
bare_get_fixed_data(struct bare_reader *ctx, uint8_t *dst, uint64_t sz)
{
	return ctx->read(ctx->buffer, dst, sz);
}

bare_error
bare_put_data(struct bare_writer *ctx, uint8_t *src, uint64_t sz)
{
	bare_error err = BARE_ERROR_NONE;

	err = bare_put_uint(ctx, sz);

	if (err == BARE_ERROR_NONE) {
		err = bare_put_fixed_data(ctx, src, sz);
	}

	return err;
}

bare_error
bare_get_data(struct bare_reader *ctx, uint8_t *dst, uint64_t sz)
{
	bare_error err = BARE_ERROR_NONE;
	uint64_t ssz = 0;

	err = bare_get_uint(ctx, &ssz);

	if (err == BARE_ERROR_NONE) {
		err = ssz <= sz \
			? bare_get_fixed_data(ctx, dst, ssz) \
			: BARE_ERROR_BUFFER_TOO_SMALL;
	}

	return err;
}

bare_error
bare_put_str(struct bare_writer *ctx, char *src, uint64_t sz)
{
	if (!checkstr(src, sz)) {
		return BARE_ERROR_INVALID_UTF8;
	}

	return bare_put_data(ctx, (uint8_t *)src, sz);
}

bare_error
bare_get_str(struct bare_reader *ctx, char *dst, uint64_t sz)
{
	bare_error err = bare_get_data(ctx, (uint8_t *)dst, sz);\

	if (err == BARE_ERROR_NONE) {
		err = !checkstr(dst, sz) ? BARE_ERROR_INVALID_UTF8 : err;
	}

	return err;
}
