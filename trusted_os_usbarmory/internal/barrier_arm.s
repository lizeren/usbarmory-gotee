//go:build tamago && arm

#include "textflag.h"

// func dsb()
// Data Synchronization Barrier (DSB SY)
TEXT ·dsb(SB),NOSPLIT,$0
	WORD	$0xf57ff04f		// DSB SY
	RET

