// Package issue193 tests for the ability to generate VPSCATTER{D,Q}{D,Q} and
// VPGATHER{D,Q}{D,Q} instructions *without* specifying a base register in
// the destination/source VSIB addressing (VM{64,32}{x,y,z}) memory operands.
package issue193

// From: https://www.felixcloutier.com/x86/vpscatterdd:vpscatterdq:vpscatterqd:vpscatterqq
//
// BASE_ADDR stands for the memory operand base address (a GPR); *may not exist*
// VINDEX stands for the memory operand vector of indices (a ZMM register)
// SCALE stands for the memory operand scalar (1, 2, 4 or 8)
// DISP is the optional 1 or 4 byte displacement
//
// For example, prior to this fix Avo requires:
//
// VPSCATTERQQ Z6, K1, 8(BX)(Z7*1)
//
// with BX = 0, when:
//
// VPSCATTERQQ Z6, K1, 8(Z7*1)
//
// is perfectly valid and will suffice.
