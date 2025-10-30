//go:build tamago && arm

package gotee

import (
	"log"

	"github.com/usbarmory/tamago/arm"
)

// Data Synchronization Barrier - ensures all memory accesses complete before proceeding
//
//go:nosplit
func dsb()

// PMU (Performance Monitoring Unit) functions for cycle-accurate timing
//
//go:nosplit
func enablePMU()

//go:nosplit
func readPMUCycleCounter() uint32

//go:nosplit
func resetPMUCycleCounter()

//go:noinline
func accessByte(ptr *byte) byte {
	return *ptr
}

// flushReload performs a Flush+Reload cache timing attack
// Returns the timing in cycles
//
//go:noinline
func flushReload(cpu *arm.CPU, ptr *byte) uint64 {
	// Step 1: FLUSH - evict the target from cache
	cpu.FlushDataCache()
	dsb()

	// Step 2: Wait for potential victim access (simulated here with delay)
	// In a real attack, victim would execute between flush and reload
	for i := 0; i < 100; i++ {
		// Busy wait
	}

	// Step 3: RELOAD - measure access time
	start := cpu.Counter()
	_ = accessByte(ptr)
	dsb()
	end := cpu.Counter()

	return end - start
}

// simulateVictimAccess simulates a victim accessing (or not accessing) memory
//
//go:noinline
func simulateVictimAccess(ptr *byte, shouldAccess bool) {
	if shouldAccess {
		_ = accessByte(ptr)
	}
}

func CacheTimerDemo() {
	log.Printf("================= Flush+Reload Cache Timing Attack Demo =================")

	cpu := arm.CPU{}
	cpu.EnableSMP()
	cpu.EnableCache()
	cpu.InitGenericTimers(0, 0)

	// Enable PMU for cycle-accurate timing
	log.Printf("\n=== Initializing Performance Monitoring Unit ===")
	enablePMU()
	resetPMUCycleCounter()

	// Test PMU resolution
	start := readPMUCycleCounter()
	dsb()
	end := readPMUCycleCounter()
	pmuOverhead := end - start
	log.Printf("PPMCCNTR: data synchronization barrier overhead: %d CPU cycles", pmuOverhead)

	// Compare with Generic Timer for reference
	gtStart := cpu.Counter()
	dsb()
	gtEnd := cpu.Counter()
	gtOverhead := gtEnd - gtStart
	log.Printf("Generic Timer: data synchronization barrier overhead: %d CPU cycles", gtOverhead)

	// Create target buffer with multiple cache lines
	// Cortex-A7 L1D cache: 32-byte lines, 256 sets, 4-way associative = 32KB total
	const (
		cacheLineSize = 32 // ARM Cortex-A7 L1D cache line size
		numLines      = 16 // Test 16 different cache lines
	)
	target := make([]byte, cacheLineSize*numLines)
	for i := range target {
		target[i] = byte(i)
	}

	log.Printf("=== Calibration: Establishing Threshold ===")

	// Calibrate: measure hit vs miss timing using PMU
	var hitSum, missSum uint64
	const calibSamples = 100

	for i := 0; i < calibSamples; i++ {
		ptr := &target[0]

		// Measure HIT using PMU
		_ = accessByte(ptr) // Prime cache
		dsb()
		start := readPMUCycleCounter()
		_ = accessByte(ptr)
		end := readPMUCycleCounter()
		hitSum += uint64(end - start)

		// Measure MISS with aggressive flushing
		cpu.FlushDataCache()
		dsb()
		start = readPMUCycleCounter()
		_ = accessByte(ptr)
		end = readPMUCycleCounter()
		missSum += uint64(end - start)
	}

	hitAvg := float64(hitSum) / float64(calibSamples)
	missAvg := float64(missSum) / float64(calibSamples)
	threshold := (hitAvg + missAvg) / 2.0

	log.Printf("Average HIT time:  %.2f CPU cycles", hitAvg)
	log.Printf("Average MISS time: %.2f CPU cycles", missAvg)
	log.Printf("Threshold: %.2f CPU cycles (midpoint)", threshold)
	log.Printf("Separation: %.2f CPU cycles (%.1fx difference)\n", missAvg-hitAvg, missAvg/hitAvg)

	log.Printf("=== Flush+Reload Attack Simulation ===")
	log.Printf("Detecting which memory locations a 'victim' accessed:\n")

	// Simulate victim accessing specific cache lines
	victimPattern := []bool{true, false, true, true, false, false, true, false,
		true, false, false, true, true, false, true, false}

	log.Printf("Victim access pattern (True=accessed, False=not accessed):")
	log.Printf("%v\n", victimPattern)

	// Attacker performs Flush+Reload on each cache line
	log.Printf("Attacker Flush+Reload measurements:")
	detected := make([]bool, numLines)

	for line := 0; line < numLines; line++ {
		ptr := &target[line*cacheLineSize] // Start of each cache line

		// FLUSH (aggressive: cache + TLB)
		cpu.FlushDataCache()
		dsb()

		// Victim accesses memory (or doesn't)
		simulateVictimAccess(ptr, victimPattern[line])

		// RELOAD and time with PMU
		start := readPMUCycleCounter()
		_ = accessByte(ptr)
		end := readPMUCycleCounter()
		timing := float64(end - start)

		// Determine if victim accessed based on timing
		wasAccessed := timing < threshold
		detected[line] = wasAccessed

		status := "MISS"
		if wasAccessed {
			status = "HIT "
		}
		log.Printf("  Line %2d: %s (%.0f CPU cycles) - detected=%v, actual=%v, %s",
			line, status, timing, wasAccessed, victimPattern[line],
			map[bool]string{true: "✓", false: "✗"}[wasAccessed == victimPattern[line]])
	}

	// Calculate accuracy
	correct := 0
	for i := 0; i < numLines; i++ {
		if detected[i] == victimPattern[i] {
			correct++
		}
	}
	accuracy := float64(correct) / float64(numLines) * 100.0

	log.Printf("\nAttack Accuracy: %d/%d (%.1f%%)", correct, numLines, accuracy)

	log.Printf("\n=== Flush+Reload Timing Distribution ===")
	log.Printf("Multiple measurements to show timing variance:\n")

	// Show timing distribution for accessed vs not-accessed using PMU
	log.Printf("Accessed (should be fast):")
	for i := 0; i < 10; i++ {
		cpu.FlushDataCache()
		dsb()
		ptr := &target[0]
		simulateVictimAccess(ptr, true)
		start := readPMUCycleCounter()
		_ = accessByte(ptr)
		end := readPMUCycleCounter()
		timing := end - start
		log.Printf("  Sample %2d: %d CPU cycles", i+1, timing)
	}

	log.Printf("\nNot Accessed (should be slow):")
	for i := 0; i < 10; i++ {
		cpu.FlushDataCache()
		dsb()
		ptr := &target[cacheLineSize] // Different cache line (line 1)
		simulateVictimAccess(ptr, false)
		start := readPMUCycleCounter()
		_ = accessByte(ptr)
		end := readPMUCycleCounter()
		timing := end - start
		log.Printf("  Sample %2d: %d CPU cycles", i+1, timing)
	}

	log.Printf("ARM Cortex-A7 L1D Cache Configuration:")
	log.Printf("  - Cache line size: 32 bytes")
	log.Printf("  - Number of sets: 256")
	log.Printf("  - Associativity: 4-way")
	log.Printf("  - Total size: 32KB (32 × 256 × 4)\n")

}
