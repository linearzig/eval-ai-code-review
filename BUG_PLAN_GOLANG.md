# GOLANG-SPECIFIC TESTING SCENARIOS

## OVERVIEW

This document outlines Golang-specific scenarios designed to test the LinearB AI Code Review Tool against language-unique characteristics, gotchas, and idioms that are PARTICULAR to Go. These scenarios focus on issues that would NOT occur in other languages and test the AI's understanding of Go's distinctive design decisions and common pitfalls.

## GOLANG-SPECIFIC CHARACTERISTICS TO TEST

### Language-Unique Features

- Zero values and nil semantics
- Interface satisfaction (implicit vs explicit)
- Method receivers (value vs pointer)
- Channel semantics and buffering
- Goroutine lifecycle management
- Context propagation patterns
- Error handling conventions (no exceptions)
- Package visibility rules
- Type embedding and composition
- Reflection limitations and patterns

### Go-Specific Gotchas

- Slices vs arrays confusion
- Map iteration randomness
- Defer execution order
- Panic/recover patterns
- Memory allocation patterns
- Garbage collection considerations
- Race conditions in concurrent code
- Channel deadlocks and leaks
- Interface nil vs concrete nil
- Type assertions and type switches

## SCENARIO COMPLEXITY LEVELS

### BEGINNER (Go-Specific Fundamentals)

Scenarios that test Go's unique syntax, zero values, and basic language constructs that differ from other languages.

### INTERMEDIATE (Go-Specific Gotchas)

Scenarios that test common Go pitfalls, interface patterns, and concurrency basics that are unique to Go.

### ADVANCED (Go-Specific Complex Patterns)

Scenarios that test advanced Go features, sophisticated concurrency patterns, and language-specific architectural decisions.

---

## PROPOSED SCENARIOS

### BEGINNER LEVEL

#### 1. **Zero Value Initialization Enhancement**

- **Camouflage**: "Add default value initialization for better safety"
- **Go-Specific Issue**: Relying on zero values when explicit initialization is needed
- **Complexity**: Low
- **Files**: Struct initialization and usage
- **Expected Catch**: Missing explicit initialization where zero values are inappropriate (e.g., time.Time{}, sync.Mutex{})

#### 2. **Interface Nil Check Enhancement**

- **Camouflage**: "Add robust interface validation"
- **Go-Specific Issue**: Interface nil vs concrete nil confusion
- **Complexity**: Low
- **Files**: Interface usage and nil checks
- **Expected Catch**: Incorrect nil checks on interfaces (interface{} vs concrete type nil)

#### 3. **Method Receiver Optimization**

- **Camouflage**: "Optimize method performance with pointer receivers"
- **Go-Specific Issue**: Value vs pointer receiver misuse
- **Complexity**: Low-Medium
- **Files**: Method definitions on structs
- **Expected Catch**: Using value receivers when pointer receivers are needed (mutating methods, large structs)

#### 4. **Slice Capacity Enhancement**

- **Camouflage**: "Add efficient slice pre-allocation"
- **Go-Specific Issue**: Slice capacity vs length confusion
- **Complexity**: Low
- **Files**: Slice operations and appending
- **Expected Catch**: Inefficient slice growth patterns, capacity vs length misuse

### INTERMEDIATE LEVEL

#### 5. **Channel Buffering Enhancement**

- **Camouflage**: "Add buffered channel support for better performance"
- **Go-Specific Issue**: Unbuffered vs buffered channel misuse
- **Complexity**: Medium
- **Files**: Channel creation and usage
- **Expected Catch**: Deadlocks from unbuffered channels, inappropriate buffer sizes

#### 6. **Goroutine Leak Prevention**

- **Camouflage**: "Add goroutine lifecycle management"
- **Go-Specific Issue**: Goroutine leaks from missing cleanup
- **Complexity**: Medium
- **Files**: Concurrent functions with goroutines
- **Expected Catch**: Goroutines that never terminate, missing done channels

#### 7. **Context Cancellation Enhancement**

- **Camouflage**: "Add request cancellation support"
- **Go-Specific Issue**: Missing context propagation and cancellation
- **Complexity**: Medium
- **Files**: HTTP handlers or long-running operations
- **Expected Catch**: Missing context.WithCancel, ignored context.Done()

#### 8. **Map Iteration Enhancement**

- **Camouflage**: "Add deterministic map processing"
- **Go-Specific Issue**: Relying on map iteration order
- **Complexity**: Medium
- **Files**: Map iteration and processing
- **Expected Catch**: Code that assumes map iteration order, missing sort for deterministic output

#### 9. **Defer Execution Enhancement**

- **Camouflage**: "Add resource cleanup management"
- **Go-Specific Issue**: Defer execution order and scope issues
- **Complexity**: Medium
- **Files**: Functions with defer statements
- **Expected Catch**: Defer in loops, defer with function calls that capture variables

#### 10. **Type Embedding Enhancement**

- **Camouflage**: "Add composition-based inheritance"
- **Go-Specific Issue**: Type embedding misuse and shadowing
- **Complexity**: Medium
- **Files**: Struct definitions with embedded types
- **Expected Catch**: Method shadowing from embedding, inappropriate type embedding

### ADVANCED LEVEL

#### 11. **Channel Select Enhancement**

- **Camouflage**: "Add multi-channel communication patterns"
- **Go-Specific Issue**: Select statement deadlocks and default cases
- **Complexity**: High
- **Files**: Complex channel communication
- **Expected Catch**: Select without default causing deadlocks, improper channel closing

#### 12. **Interface Type Assertion Enhancement**

- **Camouflage**: "Add dynamic type checking capability"
- **Go-Specific Issue**: Unsafe type assertions and type switches
- **Complexity**: High
- **Files**: Interface type checking code
- **Expected Catch**: Missing ok checks in type assertions, inefficient type switches

#### 13. **Memory Pool Enhancement**

- **Camouflage**: "Add object pooling for performance optimization"
- **Go-Specific Issue**: Inefficient memory allocation patterns
- **Complexity**: High
- **Files**: High-frequency allocation code
- **Expected Catch**: Unnecessary allocations, missing sync.Pool usage

#### 14. **Reflection Safety Enhancement**

- **Camouflage**: "Add dynamic field access capability"
- **Go-Specific Issue**: Unsafe reflection usage and performance issues
- **Complexity**: High
- **Files**: Reflection-based code
- **Expected Catch**: Reflection in hot paths, unsafe type conversions

#### 15. **Goroutine Coordination Enhancement**

- **Camouflage**: "Add sophisticated worker coordination"
- **Go-Specific Issue**: Complex goroutine synchronization and race conditions
- **Complexity**: High
- **Files**: Worker pools and complex concurrency
- **Expected Catch**: Race conditions, improper WaitGroup usage, channel leaks

---

## GO-SPECIFIC SCENARIO INTERCONNECTIONS

### **Concurrency Suite** (Scenarios 5, 6, 11, 15)

- Test Go's unique concurrency model from basic channels to complex coordination
- Progressive complexity: buffering → leaks → select → coordination

### **Interface Suite** (Scenarios 2, 10, 12)

- Test Go's interface system from nil checks to embedding to type assertions
- Progressive complexity: nil semantics → embedding → dynamic typing

### **Memory Management Suite** (Scenarios 4, 13, 14)

- Test Go's memory patterns from slices to pools to reflection
- Progressive complexity: slice efficiency → allocation patterns → reflection overhead

### **Error Handling Suite** (Scenarios 7, 9)

- Test Go's unique error handling (no exceptions) and defer patterns
- Focus on context cancellation and resource cleanup

---

## GO-SPECIFIC IMPLEMENTATION STRATEGY

### Phase 1: Go Fundamentals (Beginner Scenarios)

- Test zero values, interfaces, method receivers, and slices
- Focus on Go's unique syntax and semantics
- Build understanding of Go's design philosophy

### Phase 2: Go Gotchas (Intermediate Scenarios)

- Test channels, goroutines, context, maps, defer, and embedding
- Focus on common Go pitfalls and idioms
- Test understanding of Go's concurrency model

### Phase 3: Go Advanced Patterns (Advanced Scenarios)

- Test complex concurrency, reflection, memory management
- Focus on sophisticated Go patterns and performance considerations
- Test understanding of Go's runtime and memory model

---

## GO-SPECIFIC SCENARIO DESIGN PRINCIPLES

### Go-Unique Focus

- Every scenario must test something SPECIFIC to Go
- Avoid generic programming issues that exist in other languages
- Focus on Go's distinctive design decisions and common pitfalls

### Go Idioms and Best Practices

- Test understanding of Go's idiomatic patterns
- Focus on Go-specific performance considerations
- Test Go's unique error handling approach

### Go Runtime Understanding

- Test understanding of Go's garbage collector behavior
- Focus on Go's concurrency model and goroutine lifecycle
- Test Go's memory allocation patterns

---

## GO-SPECIFIC TESTING FOCUS AREAS

### **What Makes Go Different:**

1. **No Exceptions**: Error handling through return values
2. **Implicit Interface Satisfaction**: No explicit implements keyword
3. **Goroutines**: Lightweight threads with unique lifecycle management
4. **Channels**: First-class communication mechanism
5. **Zero Values**: Every type has a meaningful zero value
6. **Composition over Inheritance**: Type embedding and interface composition
7. **Garbage Collection**: Automatic memory management with specific patterns
8. **Package System**: Unique visibility and import rules

### **Common Go-Specific Mistakes:**

1. **Interface nil confusion**: `var i interface{} = nil` vs `var s *string = nil`
2. **Goroutine leaks**: Starting goroutines without cleanup mechanisms
3. **Channel deadlocks**: Unbuffered channels without proper coordination
4. **Slice capacity issues**: Inefficient growth patterns
5. **Method receiver confusion**: Value vs pointer receiver misuse
6. **Context misuse**: Missing cancellation or timeout
7. **Map iteration order**: Assuming deterministic iteration
8. **Defer in loops**: Capturing loop variables incorrectly

---

## IMPLEMENTATION PLAN FOR WEAVE PROJECT

### PROJECT SELECTION

**Selected Project**: Weave Net (github.com/weaveworks/weave)
**Domain**: Container networking and overlay network management
**Architecture**: Multi-host container networking with UDP encapsulation, IP allocation, DNS, and proxy services

### KEY COMPONENTS FOR SCENARIO IMPLEMENTATION

1. **Networking Layer** (`net/`): Bridge management, IP tables, network interfaces
2. **Router Layer** (`router/`): Packet forwarding, encryption, UDP communication
3. **Common Utilities** (`common/`): Logging, HTTP handling, process management
4. **IPAM** (`ipam/`): IP address management and allocation
5. **Nameserver** (`nameserver/`): DNS resolution services
6. **Proxy** (`proxy/`): HTTP proxy for container communication

---

## DETAILED SCENARIO IMPLEMENTATION MAPPING

### BEGINNER LEVEL SCENARIOS

#### 1. **Zero Value Initialization Enhancement**

- **Target File**: `net/bridge.go` - BridgeConfig struct initialization
- **Camouflage**: "Add default bridge configuration for better safety"
- **Go-Specific Issue**: Relying on zero values for time.Time{}, sync.Mutex{}, or complex structs
- **Implementation**: In `BridgeConfig` struct or bridge initialization functions
- **Expected Catch**: Missing explicit initialization of time fields, mutexes, or complex nested structs
- **Scenario Name**: `zero-value-initialization`

#### 2. **Interface Nil Check Enhancement**

- **Target File**: `router/sleeve.go` - Interface nil vs concrete nil confusion
- **Camouflage**: "Add robust interface validation for packet processing"
- **Go-Specific Issue**: Interface nil vs concrete nil confusion in packet handlers
- **Implementation**: In `SleeveOverlay` struct or packet processing functions
- **Expected Catch**: Incorrect nil checks on interface{} vs concrete type nil
- **Scenario Name**: `interface-nil-check`

#### 3. **Method Receiver Optimization**

- **Target File**: `net/bridge.go` - Bridge interface implementations
- **Camouflage**: "Optimize bridge method performance with pointer receivers"
- **Go-Specific Issue**: Value vs pointer receiver misuse in bridge methods
- **Implementation**: In `bridgeImpl`, `fastdpImpl`, or `bridgedFastdpImpl` methods
- **Expected Catch**: Using value receivers when pointer receivers are needed for mutating methods
- **Scenario Name**: `method-receiver-optimization`

#### 4. **Slice Capacity Enhancement**

- **Target File**: `router/sleeve.go` - Packet aggregation and buffering
- **Camouflage**: "Add efficient packet buffer pre-allocation"
- **Go-Specific Issue**: Slice capacity vs length confusion in packet buffers
- **Implementation**: In `sleeveForwarder` packet aggregation functions
- **Expected Catch**: Inefficient slice growth patterns, capacity vs length misuse
- **Scenario Name**: `slice-capacity-enhancement`

### INTERMEDIATE LEVEL SCENARIOS

#### 5. **Channel Buffering Enhancement**

- **Target File**: `router/sleeve.go` - Channel communication between goroutines
- **Camouflage**: "Add buffered channel support for better packet throughput"
- **Go-Specific Issue**: Unbuffered vs buffered channel misuse causing deadlocks
- **Implementation**: In `sleeveForwarder` channel creation (aggregatorChan, controlMsgChan)
- **Expected Catch**: Deadlocks from unbuffered channels, inappropriate buffer sizes
- **Scenario Name**: `channel-buffering-enhancement`

#### 6. **Goroutine Leak Prevention**

- **Target File**: `router/sleeve.go` - Goroutine lifecycle management
- **Camouflage**: "Add goroutine lifecycle management for packet forwarders"
- **Go-Specific Issue**: Goroutine leaks from missing cleanup in packet forwarders
- **Implementation**: In `sleeveForwarder.run()` method or forwarder creation
- **Expected Catch**: Goroutines that never terminate, missing done channels
- **Scenario Name**: `goroutine-leak-prevention`

#### 7. **Context Cancellation Enhancement**

- **Target File**: `common/http.go` or `router/http.go` - HTTP handlers
- **Camouflage**: "Add request cancellation support for HTTP endpoints"
- **Go-Specific Issue**: Missing context propagation and cancellation in HTTP handlers
- **Implementation**: In HTTP handler functions or API endpoints
- **Expected Catch**: Missing context.WithCancel, ignored context.Done()
- **Scenario Name**: `context-cancellation-enhancement`

#### 8. **Map Iteration Enhancement**

- **Target File**: `router/mac_cache.go` - MAC address caching
- **Camouflage**: "Add deterministic MAC cache processing"
- **Go-Specific Issue**: Relying on map iteration order in MAC cache operations
- **Implementation**: In MAC cache iteration and processing functions
- **Expected Catch**: Code that assumes map iteration order, missing sort for deterministic output
- **Scenario Name**: `map-iteration-enhancement`

#### 9. **Defer Execution Enhancement**

- **Target File**: `net/bridge.go` - Resource cleanup management
- **Camouflage**: "Add resource cleanup management for bridge operations"
- **Go-Specific Issue**: Defer execution order and scope issues in bridge setup
- **Implementation**: In bridge initialization or cleanup functions
- **Expected Catch**: Defer in loops, defer with function calls that capture variables
- **Scenario Name**: `defer-execution-enhancement`

#### 10. **Type Embedding Enhancement**

- **Target File**: `net/bridge.go` - Bridge type composition
- **Camouflage**: "Add composition-based bridge inheritance"
- **Go-Specific Issue**: Type embedding misuse and method shadowing
- **Implementation**: In `bridgedFastdpImpl` struct with embedded types
- **Expected Catch**: Method shadowing from embedding, inappropriate type embedding
- **Scenario Name**: `type-embedding-enhancement`

### ADVANCED LEVEL SCENARIOS

#### 11. **Channel Select Enhancement**

- **Target File**: `router/sleeve.go` - Multi-channel communication patterns
- **Camouflage**: "Add multi-channel communication patterns for packet routing"
- **Go-Specific Issue**: Select statement deadlocks and missing default cases
- **Implementation**: In `sleeveForwarder.run()` method with multiple channels
- **Expected Catch**: Select without default causing deadlocks, improper channel closing
- **Scenario Name**: `channel-select-enhancement`

#### 12. **Interface Type Assertion Enhancement**

- **Target File**: `router/ethernet_decoder.go` - Dynamic type checking
- **Camouflage**: "Add dynamic type checking capability for packet decoding"
- **Go-Specific Issue**: Unsafe type assertions and type switches
- **Implementation**: In packet decoding or type checking functions
- **Expected Catch**: Missing ok checks in type assertions, inefficient type switches
- **Scenario Name**: `interface-type-assertion-enhancement`

#### 13. **Memory Pool Enhancement**

- **Target File**: `router/sleeve.go` - High-frequency packet allocation
- **Camouflage**: "Add object pooling for packet buffer optimization"
- **Go-Specific Issue**: Inefficient memory allocation patterns in packet processing
- **Implementation**: In packet buffer allocation and reuse
- **Expected Catch**: Unnecessary allocations, missing sync.Pool usage
- **Scenario Name**: `memory-pool-enhancement`

#### 14. **Reflection Safety Enhancement**

- **Target File**: `common/` or configuration handling - Dynamic field access
- **Camouflage**: "Add dynamic field access capability for configuration"
- **Go-Specific Issue**: Unsafe reflection usage and performance issues
- **Implementation**: In configuration parsing or dynamic field access
- **Expected Catch**: Reflection in hot paths, unsafe type conversions
- **Scenario Name**: `reflection-safety-enhancement`

#### 15. **Goroutine Coordination Enhancement**

- **Target File**: `router/sleeve.go` - Complex goroutine synchronization
- **Camouflage**: "Add sophisticated worker coordination for packet processing"
- **Go-Specific Issue**: Complex goroutine synchronization and race conditions
- **Implementation**: In packet forwarder coordination and worker pools
- **Expected Catch**: Race conditions, improper WaitGroup usage, channel leaks
- **Scenario Name**: `goroutine-coordination-enhancement`

---

## IMPLEMENTATION STRATEGY

### Phase 1: Foundation (Beginner Scenarios 1-4)

**Focus**: Basic Go syntax and semantics in networking code

- **Files**: `net/bridge.go`, `router/sleeve.go`
- **Goal**: Test AI understanding of Go's zero values, interfaces, method receivers, and slices
- **Complexity**: Low - fundamental Go concepts

### Phase 2: Networking Gotchas (Intermediate Scenarios 5-10)

**Focus**: Go-specific pitfalls in networking and concurrency

- **Files**: `router/sleeve.go`, `net/bridge.go`, `router/mac_cache.go`
- **Goal**: Test AI understanding of channels, goroutines, context, maps, defer, and embedding
- **Complexity**: Medium - common Go networking patterns

### Phase 3: Advanced Networking Patterns (Advanced Scenarios 11-15)

**Focus**: Complex Go patterns in high-performance networking

- **Files**: `router/sleeve.go`, `router/ethernet_decoder.go`, `common/`
- **Goal**: Test AI understanding of advanced concurrency, reflection, memory management
- **Complexity**: High - sophisticated Go networking patterns

---

## SCENARIO-SPECIFIC IMPLEMENTATION DETAILS

### **Scenario 1: Zero Value Initialization**

```go
// In net/bridge.go - BridgeConfig struct
type BridgeConfig struct {
    // ... existing fields ...
    LastUpdated time.Time  // Zero value issue: should be explicitly initialized
    Mutex       sync.Mutex // Zero value issue: should be explicitly initialized
}
```

### **Scenario 5: Channel Buffering**

```go
// In router/sleeve.go - sleeveForwarder struct
type sleeveForwarder struct {
    // ... existing fields ...
    aggregatorChan   chan<- aggregatorFrame  // Unbuffered - potential deadlock
    controlMsgChan   chan<- controlMessage   // Unbuffered - potential deadlock
}
```

### **Scenario 8: Map Iteration**

```go
// In router/mac_cache.go - MAC cache processing
func (cache *MACCache) ProcessAll() {
    for mac, entry := range cache.entries { // Assumes iteration order
        // Process entry
    }
}
```

### **Scenario 11: Channel Select**

```go
// In router/sleeve.go - sleeveForwarder.run()
func (fwd *sleeveForwarder) run(...) {
    for {
        select {
        case frame := <-aggChan:
            // Process frame
        case msg := <-controlMsgChan:
            // Process message
        // Missing default case - potential deadlock
        }
    }
}
```

---

## TESTING APPROACH

### **Realistic Feature Additions**

Each scenario will be implemented as a legitimate-looking feature addition:

- Configuration enhancements
- Performance optimizations
- New capabilities
- Bug fixes (that introduce new bugs)

### **Camouflage Techniques**

- Use realistic feature names and descriptions
- Follow existing code patterns and conventions
- Add appropriate comments and documentation
- Maintain code style consistency

### **Expected AI Review Outcomes**

- Catch Go-specific issues that would be missed in other languages
- Identify concurrency problems unique to Go
- Spot interface and type system issues
- Recognize memory management patterns

---

## SUCCESS CRITERIA

### **Go-Specific Issue Detection**

- AI should catch issues that are UNIQUE to Go
- Should understand Go's concurrency model
- Should recognize Go's type system quirks
- Should identify Go-specific performance patterns

### **False Positive Management**

- Scenarios should not trigger false positives for legitimate Go patterns
- AI should understand Go idioms and best practices
- Should distinguish between actual issues and Go's design decisions

### **Comprehensive Coverage**

- Test all major Go-specific characteristics
- Cover beginner to advanced complexity levels
- Include both static and runtime issues
- Test understanding of Go's philosophy and patterns

---

## NEXT STEPS

1. **Create Scenario Description Files**: Write `.txt` files for each scenario
2. **Implement Scenario Code**: Create actual code changes for each scenario
3. **Test Implementation**: Validate scenarios against the AI review tool
4. **Document Results**: Track which Go-specific issues are effectively caught
5. **Refine and Iterate**: Improve scenarios based on testing results

---

## NOTES

- Focus on issues that are UNIQUE to Go and would not occur in other languages
- Test the AI's understanding of Go's design philosophy and idioms
- Emphasize Go's concurrency model and memory management patterns
- Consider Go's performance characteristics and best practices
- Test understanding of Go's type system and interface patterns
- All scenarios target the Weave Net project for realistic testing
