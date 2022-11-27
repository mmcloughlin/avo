// Code generated by command: avogen -output zmov.go mov. DO NOT EDIT.

package build

import (
	"go/types"

	"github.com/mmcloughlin/avo/operand"
)

func (c *Context) mov(a, b operand.Op, an, bn int, t *types.Basic) {
	switch {
	case an == 8 && operand.IsK(a) && bn == 8 && operand.IsK(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.KMOVB(a, b)
	case an == 8 && operand.IsK(a) && bn == 1 && operand.IsM8(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.KMOVB(a, b)
	case an == 8 && operand.IsK(a) && bn == 4 && operand.IsR32(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.KMOVB(a, b)
	case an == 1 && operand.IsM8(a) && bn == 8 && operand.IsK(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.KMOVB(a, b)
	case an == 4 && operand.IsR32(a) && bn == 8 && operand.IsK(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.KMOVB(a, b)
	case an == 8 && operand.IsK(a) && bn == 8 && operand.IsK(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.KMOVD(a, b)
	case an == 8 && operand.IsK(a) && bn == 4 && operand.IsM32(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.KMOVD(a, b)
	case an == 8 && operand.IsK(a) && bn == 4 && operand.IsR32(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.KMOVD(a, b)
	case an == 4 && operand.IsM32(a) && bn == 8 && operand.IsK(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.KMOVD(a, b)
	case an == 4 && operand.IsR32(a) && bn == 8 && operand.IsK(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.KMOVD(a, b)
	case an == 8 && operand.IsK(a) && bn == 8 && operand.IsK(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.KMOVQ(a, b)
	case an == 8 && operand.IsK(a) && bn == 8 && operand.IsM64(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.KMOVQ(a, b)
	case an == 8 && operand.IsK(a) && bn == 8 && operand.IsR64(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.KMOVQ(a, b)
	case an == 8 && operand.IsM64(a) && bn == 8 && operand.IsK(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.KMOVQ(a, b)
	case an == 8 && operand.IsR64(a) && bn == 8 && operand.IsK(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.KMOVQ(a, b)
	case an == 8 && operand.IsK(a) && bn == 8 && operand.IsK(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.KMOVW(a, b)
	case an == 8 && operand.IsK(a) && bn == 2 && operand.IsM16(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.KMOVW(a, b)
	case an == 8 && operand.IsK(a) && bn == 4 && operand.IsR32(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.KMOVW(a, b)
	case an == 2 && operand.IsM16(a) && bn == 8 && operand.IsK(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.KMOVW(a, b)
	case an == 4 && operand.IsR32(a) && bn == 8 && operand.IsK(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.KMOVW(a, b)
	case an == 1 && operand.IsM8(a) && bn == 1 && operand.IsR8(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVB(a, b)
	case an == 1 && operand.IsR8(a) && bn == 1 && operand.IsM8(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVB(a, b)
	case an == 1 && operand.IsR8(a) && bn == 1 && operand.IsR8(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVB(a, b)
	case an == 1 && operand.IsM8(a) && bn == 4 && operand.IsR32(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == types.IsInteger:
		c.MOVBLSX(a, b)
	case an == 1 && operand.IsR8(a) && bn == 4 && operand.IsR32(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == types.IsInteger:
		c.MOVBLSX(a, b)
	case an == 1 && operand.IsM8(a) && bn == 4 && operand.IsR32(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == (types.IsInteger|types.IsUnsigned):
		c.MOVBLZX(a, b)
	case an == 1 && operand.IsM8(a) && bn == 4 && operand.IsR32(b) && (t.Info()&types.IsBoolean) == types.IsBoolean:
		c.MOVBLZX(a, b)
	case an == 1 && operand.IsR8(a) && bn == 4 && operand.IsR32(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == (types.IsInteger|types.IsUnsigned):
		c.MOVBLZX(a, b)
	case an == 1 && operand.IsR8(a) && bn == 4 && operand.IsR32(b) && (t.Info()&types.IsBoolean) == types.IsBoolean:
		c.MOVBLZX(a, b)
	case an == 1 && operand.IsM8(a) && bn == 8 && operand.IsR64(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == types.IsInteger:
		c.MOVBQSX(a, b)
	case an == 1 && operand.IsR8(a) && bn == 8 && operand.IsR64(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == types.IsInteger:
		c.MOVBQSX(a, b)
	case an == 1 && operand.IsM8(a) && bn == 8 && operand.IsR64(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == (types.IsInteger|types.IsUnsigned):
		c.MOVBQZX(a, b)
	case an == 1 && operand.IsM8(a) && bn == 8 && operand.IsR64(b) && (t.Info()&types.IsBoolean) == types.IsBoolean:
		c.MOVBQZX(a, b)
	case an == 1 && operand.IsR8(a) && bn == 8 && operand.IsR64(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == (types.IsInteger|types.IsUnsigned):
		c.MOVBQZX(a, b)
	case an == 1 && operand.IsR8(a) && bn == 8 && operand.IsR64(b) && (t.Info()&types.IsBoolean) == types.IsBoolean:
		c.MOVBQZX(a, b)
	case an == 1 && operand.IsM8(a) && bn == 2 && operand.IsR16(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == types.IsInteger:
		c.MOVBWSX(a, b)
	case an == 1 && operand.IsR8(a) && bn == 2 && operand.IsR16(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == types.IsInteger:
		c.MOVBWSX(a, b)
	case an == 1 && operand.IsM8(a) && bn == 2 && operand.IsR16(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == (types.IsInteger|types.IsUnsigned):
		c.MOVBWZX(a, b)
	case an == 1 && operand.IsM8(a) && bn == 2 && operand.IsR16(b) && (t.Info()&types.IsBoolean) == types.IsBoolean:
		c.MOVBWZX(a, b)
	case an == 1 && operand.IsR8(a) && bn == 2 && operand.IsR16(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == (types.IsInteger|types.IsUnsigned):
		c.MOVBWZX(a, b)
	case an == 1 && operand.IsR8(a) && bn == 2 && operand.IsR16(b) && (t.Info()&types.IsBoolean) == types.IsBoolean:
		c.MOVBWZX(a, b)
	case an == 4 && operand.IsM32(a) && bn == 4 && operand.IsR32(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVL(a, b)
	case an == 4 && operand.IsR32(a) && bn == 4 && operand.IsM32(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVL(a, b)
	case an == 4 && operand.IsR32(a) && bn == 4 && operand.IsR32(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVL(a, b)
	case an == 4 && operand.IsM32(a) && bn == 8 && operand.IsR64(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == types.IsInteger:
		c.MOVLQSX(a, b)
	case an == 4 && operand.IsR32(a) && bn == 8 && operand.IsR64(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == types.IsInteger:
		c.MOVLQSX(a, b)
	case an == 4 && operand.IsM32(a) && bn == 8 && operand.IsR64(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == (types.IsInteger|types.IsUnsigned):
		c.MOVLQZX(a, b)
	case an == 4 && operand.IsM32(a) && bn == 8 && operand.IsR64(b) && (t.Info()&types.IsBoolean) == types.IsBoolean:
		c.MOVLQZX(a, b)
	case an == 16 && operand.IsM128(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVOU(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 16 && operand.IsM128(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVOU(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVOU(a, b)
	case an == 4 && operand.IsM32(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVQ(a, b)
	case an == 8 && operand.IsM64(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVQ(a, b)
	case an == 4 && operand.IsR32(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVQ(a, b)
	case an == 8 && operand.IsR64(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVQ(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 4 && operand.IsM32(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVQ(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 8 && operand.IsM64(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVQ(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 4 && operand.IsR32(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVQ(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 8 && operand.IsR64(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVQ(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVQ(a, b)
	case an == 8 && operand.IsM64(a) && bn == 8 && operand.IsR64(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVQ(a, b)
	case an == 8 && operand.IsR64(a) && bn == 8 && operand.IsM64(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVQ(a, b)
	case an == 8 && operand.IsR64(a) && bn == 8 && operand.IsR64(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVQ(a, b)
	case an == 8 && operand.IsM64(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsFloat) == types.IsFloat:
		c.MOVSD(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 8 && operand.IsM64(b) && (t.Info()&types.IsFloat) == types.IsFloat:
		c.MOVSD(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsFloat) == types.IsFloat:
		c.MOVSD(a, b)
	case an == 4 && operand.IsM32(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsFloat) == types.IsFloat:
		c.MOVSS(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 4 && operand.IsM32(b) && (t.Info()&types.IsFloat) == types.IsFloat:
		c.MOVSS(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsFloat) == types.IsFloat:
		c.MOVSS(a, b)
	case an == 2 && operand.IsM16(a) && bn == 2 && operand.IsR16(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVW(a, b)
	case an == 2 && operand.IsR16(a) && bn == 2 && operand.IsM16(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVW(a, b)
	case an == 2 && operand.IsR16(a) && bn == 2 && operand.IsR16(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.MOVW(a, b)
	case an == 2 && operand.IsM16(a) && bn == 4 && operand.IsR32(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == types.IsInteger:
		c.MOVWLSX(a, b)
	case an == 2 && operand.IsR16(a) && bn == 4 && operand.IsR32(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == types.IsInteger:
		c.MOVWLSX(a, b)
	case an == 2 && operand.IsM16(a) && bn == 4 && operand.IsR32(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == (types.IsInteger|types.IsUnsigned):
		c.MOVWLZX(a, b)
	case an == 2 && operand.IsM16(a) && bn == 4 && operand.IsR32(b) && (t.Info()&types.IsBoolean) == types.IsBoolean:
		c.MOVWLZX(a, b)
	case an == 2 && operand.IsR16(a) && bn == 4 && operand.IsR32(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == (types.IsInteger|types.IsUnsigned):
		c.MOVWLZX(a, b)
	case an == 2 && operand.IsR16(a) && bn == 4 && operand.IsR32(b) && (t.Info()&types.IsBoolean) == types.IsBoolean:
		c.MOVWLZX(a, b)
	case an == 2 && operand.IsM16(a) && bn == 8 && operand.IsR64(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == types.IsInteger:
		c.MOVWQSX(a, b)
	case an == 2 && operand.IsR16(a) && bn == 8 && operand.IsR64(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == types.IsInteger:
		c.MOVWQSX(a, b)
	case an == 2 && operand.IsM16(a) && bn == 8 && operand.IsR64(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == (types.IsInteger|types.IsUnsigned):
		c.MOVWQZX(a, b)
	case an == 2 && operand.IsM16(a) && bn == 8 && operand.IsR64(b) && (t.Info()&types.IsBoolean) == types.IsBoolean:
		c.MOVWQZX(a, b)
	case an == 2 && operand.IsR16(a) && bn == 8 && operand.IsR64(b) && (t.Info()&(types.IsInteger|types.IsUnsigned)) == (types.IsInteger|types.IsUnsigned):
		c.MOVWQZX(a, b)
	case an == 2 && operand.IsR16(a) && bn == 8 && operand.IsR64(b) && (t.Info()&types.IsBoolean) == types.IsBoolean:
		c.MOVWQZX(a, b)
	case an == 4 && operand.IsM32(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVD(a, b)
	case an == 4 && operand.IsR32(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVD(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 4 && operand.IsM32(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVD(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 4 && operand.IsR32(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVD(a, b)
	case an == 16 && operand.IsM128(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU(a, b)
	case an == 32 && operand.IsM256(a) && bn == 32 && operand.IsYMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 16 && operand.IsM128(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU(a, b)
	case an == 32 && operand.IsYMM(a) && bn == 32 && operand.IsM256(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU(a, b)
	case an == 32 && operand.IsYMM(a) && bn == 32 && operand.IsYMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU(a, b)
	case an == 16 && operand.IsM128(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU16(a, b)
	case an == 32 && operand.IsM256(a) && bn == 32 && operand.IsYMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU16(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 16 && operand.IsM128(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU16(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU16(a, b)
	case an == 32 && operand.IsYMM(a) && bn == 32 && operand.IsM256(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU16(a, b)
	case an == 32 && operand.IsYMM(a) && bn == 32 && operand.IsYMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU16(a, b)
	case an == 64 && operand.IsM512(a) && bn == 64 && operand.IsZMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU16(a, b)
	case an == 64 && operand.IsZMM(a) && bn == 64 && operand.IsM512(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU16(a, b)
	case an == 64 && operand.IsZMM(a) && bn == 64 && operand.IsZMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU16(a, b)
	case an == 16 && operand.IsM128(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU32(a, b)
	case an == 32 && operand.IsM256(a) && bn == 32 && operand.IsYMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU32(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 16 && operand.IsM128(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU32(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU32(a, b)
	case an == 32 && operand.IsYMM(a) && bn == 32 && operand.IsM256(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU32(a, b)
	case an == 32 && operand.IsYMM(a) && bn == 32 && operand.IsYMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU32(a, b)
	case an == 64 && operand.IsM512(a) && bn == 64 && operand.IsZMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU32(a, b)
	case an == 64 && operand.IsZMM(a) && bn == 64 && operand.IsM512(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU32(a, b)
	case an == 64 && operand.IsZMM(a) && bn == 64 && operand.IsZMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU32(a, b)
	case an == 16 && operand.IsM128(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU64(a, b)
	case an == 32 && operand.IsM256(a) && bn == 32 && operand.IsYMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU64(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 16 && operand.IsM128(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU64(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU64(a, b)
	case an == 32 && operand.IsYMM(a) && bn == 32 && operand.IsM256(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU64(a, b)
	case an == 32 && operand.IsYMM(a) && bn == 32 && operand.IsYMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU64(a, b)
	case an == 64 && operand.IsM512(a) && bn == 64 && operand.IsZMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU64(a, b)
	case an == 64 && operand.IsZMM(a) && bn == 64 && operand.IsM512(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU64(a, b)
	case an == 64 && operand.IsZMM(a) && bn == 64 && operand.IsZMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU64(a, b)
	case an == 16 && operand.IsM128(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU8(a, b)
	case an == 32 && operand.IsM256(a) && bn == 32 && operand.IsYMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU8(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 16 && operand.IsM128(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU8(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU8(a, b)
	case an == 32 && operand.IsYMM(a) && bn == 32 && operand.IsM256(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU8(a, b)
	case an == 32 && operand.IsYMM(a) && bn == 32 && operand.IsYMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU8(a, b)
	case an == 64 && operand.IsM512(a) && bn == 64 && operand.IsZMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU8(a, b)
	case an == 64 && operand.IsZMM(a) && bn == 64 && operand.IsM512(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU8(a, b)
	case an == 64 && operand.IsZMM(a) && bn == 64 && operand.IsZMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVDQU8(a, b)
	case an == 8 && operand.IsM64(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVQ(a, b)
	case an == 8 && operand.IsR64(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVQ(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 8 && operand.IsM64(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVQ(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 8 && operand.IsR64(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVQ(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsInteger) == types.IsInteger:
		c.VMOVQ(a, b)
	case an == 8 && operand.IsM64(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsFloat) == types.IsFloat:
		c.VMOVSD(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 8 && operand.IsM64(b) && (t.Info()&types.IsFloat) == types.IsFloat:
		c.VMOVSD(a, b)
	case an == 4 && operand.IsM32(a) && bn == 16 && operand.IsXMM(b) && (t.Info()&types.IsFloat) == types.IsFloat:
		c.VMOVSS(a, b)
	case an == 16 && operand.IsXMM(a) && bn == 4 && operand.IsM32(b) && (t.Info()&types.IsFloat) == types.IsFloat:
		c.VMOVSS(a, b)
	default:
		c.adderrormessage("could not deduce mov instruction")
	}
}
