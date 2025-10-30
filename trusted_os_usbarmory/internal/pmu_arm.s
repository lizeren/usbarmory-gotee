//go:build tamago && arm

#include "textflag.h"

// func enablePMU()
// Enable Performance Monitoring Unit
TEXT ·enablePMU(SB),NOSPLIT,$0
	// Enable PMU user-mode access (PMUSERENR)
	MOVW	$1, R0
	MCR	15, 0, R0, C9, C14, 0
	
	// Enable all counters (PMCR)
	MRC	15, 0, R0, C9, C12, 0
	ORR	$1, R0              // Enable all counters
	ORR	$(1<<2), R0         // Reset cycle counter
	ORR	$(1<<3), R0         // Reset all counters
	MCR	15, 0, R0, C9, C12, 0
	
	// Enable cycle counter (PMCNTENSET)
	MOVW	$(1<<31), R0        // Enable cycle counter (bit 31)
	MCR	15, 0, R0, C9, C12, 1
	
	RET

// func readPMUCycleCounter() uint32
// Read PMU cycle counter (PMCCNTR)
TEXT ·readPMUCycleCounter(SB),NOSPLIT,$0-4
	MRC	15, 0, R0, C9, C13, 0
	MOVW	R0, ret+0(FP)
	RET

// func resetPMUCycleCounter()
// Reset PMU cycle counter
TEXT ·resetPMUCycleCounter(SB),NOSPLIT,$0
	MRC	15, 0, R0, C9, C12, 0
	ORR	$(1<<2), R0         // Reset cycle counter
	MCR	15, 0, R0, C9, C12, 0
	RET
