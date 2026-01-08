# Examples

This section provides practical examples of using grule-plus in various scenarios.

## Basic Usage

### Simple Rule Engine

```go
package main

import (
    "context"
    "fmt"
    "github.com/hungpdn/grule-plus/engine"
)

func main() {
    // Configure the engine
    cfg := engine.Config{
        Type:            engine.LRU,
        Size:            1000,
        CleanupInterval: 10,
        TTL:             60,
        Partition:       1,
        FactName:        "DiscountFact",
    }

    // Create engine instance
    grule := engine.NewPartitionEngine(cfg, nil)
    defer grule.Close()

    // Define a rule
    rule := "DiscountRule"
    statement := `rule DiscountRule "Apply discount" salience 10 {
        when
            DiscountFact.Amount > 100
        then
            DiscountFact.Discount = 10;
            Retract("DiscountRule");
    }`

    // Add the rule
    err := grule.AddRule(rule, statement, 60)
    if err != nil {
        panic(err)
    }

    // Create fact
    fact := struct {
        Amount   int
        Discount int
    }{Amount: 150}

    // Execute rule
    err = grule.Execute(context.Background(), rule, &fact)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Final discount: %d%%\n", fact.Discount) // Output: 10%
}
```

## Advanced Examples

### E-commerce Discount Rules

```go
package main

import (
    "context"
    "fmt"
    "github.com/hungpdn/grule-plus/engine"
)

type Customer struct {
    LoyaltyLevel string
    TotalSpent   float64
}

type Product struct {
    Category string
    Price    float64
}

type DiscountContext struct {
    Customer *Customer
    Product  *Product
    Discount float64
}

func main() {
    cfg := engine.Config{
        Type:            engine.ARC,
        Size:            1000,
        CleanupInterval: 30,
        TTL:             300,
        Partition:       4,
        FactName:        "DiscountContext",
    }

    grule := engine.NewPartitionEngine(cfg, nil)
    defer grule.Close()

    // VIP customer discount
    vipRule := `rule VIPDiscount "VIP customer discount" salience 100 {
        when
            DiscountContext.Customer.LoyaltyLevel == "VIP" &&
            DiscountContext.Product.Price > 50
        then
            DiscountContext.Discount = 20.0;
            Retract("VIPDiscount");
    }`

    // High spender discount
    spenderRule := `rule SpenderDiscount "High spender discount" salience 50 {
        when
            DiscountContext.Customer.TotalSpent > 1000 &&
            DiscountContext.Product.Category == "Electronics"
        then
            DiscountContext.Discount = 15.0;
            Retract("SpenderDiscount");
    }`

    // Add rules
    grule.AddRule("VIPDiscount", vipRule, 300)
    grule.AddRule("SpenderDiscount", spenderRule, 300)

    // Test cases
    testCases := []DiscountContext{
        {
            Customer: &Customer{LoyaltyLevel: "VIP", TotalSpent: 500},
            Product:  &Product{Category: "Electronics", Price: 100},
            Discount: 0,
        },
        {
            Customer: &Customer{LoyaltyLevel: "Regular", TotalSpent: 1200},
            Product:  &Product{Category: "Electronics", Price: 80},
            Discount: 0,
        },
    }

    for i, tc := range testCases {
        // Try VIP rule first
        grule.Execute(context.Background(), "VIPDiscount", &tc)

        // If no VIP discount, try spender rule
        if tc.Discount == 0 {
            grule.Execute(context.Background(), "SpenderDiscount", &tc)
        }

        fmt.Printf("Test case %d: Discount %.1f%%\n", i+1, tc.Discount)
    }
}
```

### Financial Risk Assessment

```go
package main

import (
    "context"
    "fmt"
    "github.com/hungpdn/grule-plus/engine"
)

type Transaction struct {
    Amount      float64
    Merchant    string
    Country     string
    CardType    string
    RiskScore   float64
}

type RiskAssessment struct {
    Transaction *Transaction
    RiskLevel   string
    Approved    bool
}

func main() {
    cfg := engine.Config{
        Type:            engine.LFU,
        Size:            2000,
        CleanupInterval: 60,
        TTL:             600,
        Partition:       2,
        FactName:        "RiskAssessment",
    }

    grule := engine.NewPartitionEngine(cfg, nil)
    defer grule.Close()

    // High amount rule
    highAmountRule := `rule HighAmountCheck "Check high transaction amounts" salience 100 {
        when
            RiskAssessment.Transaction.Amount > 10000
        then
            RiskAssessment.RiskScore += 30;
            Retract("HighAmountCheck");
    }`

    // International transaction rule
    intlRule := `rule InternationalCheck "Check international transactions" salience 80 {
        when
            RiskAssessment.Transaction.Country != "US"
        then
            RiskAssessment.RiskScore += 20;
            Retract("InternationalCheck");
    }`

    // Risk level determination
    riskLevelRule := `rule RiskLevelDetermination "Determine risk level" salience 10 {
        when
            RiskAssessment.RiskScore >= 40
        then
            RiskAssessment.RiskLevel = "HIGH";
            RiskAssessment.Approved = false;
            Retract("RiskLevelDetermination");
    }`

    // Add rules
    grule.AddRule("HighAmountCheck", highAmountRule, 600)
    grule.AddRule("InternationalCheck", intlRule, 600)
    grule.AddRule("RiskLevelDetermination", riskLevelRule, 600)

    // Test transaction
    assessment := RiskAssessment{
        Transaction: &Transaction{
            Amount:    15000,
            Merchant:  "OnlineStore",
            Country:   "CA",
            CardType:  "Credit",
            RiskScore: 0,
        },
        RiskLevel: "LOW",
        Approved:  true,
    }

    // Execute all applicable rules
    grule.Execute(context.Background(), "HighAmountCheck", &assessment)
    grule.Execute(context.Background(), "InternationalCheck", &assessment)
    grule.Execute(context.Background(), "RiskLevelDetermination", &assessment)

    fmt.Printf("Transaction Risk: %s, Approved: %t, Score: %.0f\n",
        assessment.RiskLevel, assessment.Approved, assessment.Transaction.RiskScore)
}
```

### IoT Device Monitoring

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/hungpdn/grule-plus/engine"
)

type SensorData struct {
    DeviceID    string
    Temperature float64
    Humidity    float64
    Timestamp   time.Time
}

type AlertContext struct {
    Sensor   *SensorData
    Alert    bool
    Severity string
    Message  string
}

func main() {
    cfg := engine.Config{
        Type:            engine.TWOQ,
        Size:            5000,
        CleanupInterval: 30,
        TTL:             1800, // 30 minutes
        Partition:       8,
        FactName:        "AlertContext",
    }

    grule := engine.NewPartitionEngine(cfg, nil)
    defer grule.Close()

    // Temperature alert rule
    tempRule := `rule TemperatureAlert "Monitor temperature thresholds" salience 50 {
        when
            AlertContext.Sensor.Temperature > 35.0 ||
            AlertContext.Sensor.Temperature < 5.0
        then
            AlertContext.Alert = true;
            AlertContext.Severity = "WARNING";
            AlertContext.Message = "Temperature out of range";
            Retract("TemperatureAlert");
    }`

    // Humidity alert rule
    humidityRule := `rule HumidityAlert "Monitor humidity levels" salience 40 {
        when
            AlertContext.Sensor.Humidity > 80.0 ||
            AlertContext.Sensor.Humidity < 20.0
        then
            AlertContext.Alert = true;
            if AlertContext.Severity == "" {
                AlertContext.Severity = "INFO";
            }
            AlertContext.Message = AlertContext.Message + " Humidity abnormal";
            Retract("HumidityAlert");
    }`

    // Critical temperature rule
    criticalRule := `rule CriticalTemperature "Critical temperature alert" salience 100 {
        when
            AlertContext.Sensor.Temperature > 45.0 ||
            AlertContext.Sensor.Temperature < 0.0
        then
            AlertContext.Alert = true;
            AlertContext.Severity = "CRITICAL";
            AlertContext.Message = "CRITICAL: Temperature extreme";
            Retract("CriticalTemperature");
    }`

    // Add rules with longer TTL for IoT rules
    grule.AddRule("TemperatureAlert", tempRule, 1800)
    grule.AddRule("HumidityAlert", humidityRule, 1800)
    grule.AddRule("CriticalTemperature", criticalRule, 1800)

    // Simulate sensor readings
    readings := []SensorData{
        {DeviceID: "sensor1", Temperature: 38.5, Humidity: 65.0, Timestamp: time.Now()},
        {DeviceID: "sensor2", Temperature: 2.0, Humidity: 85.0, Timestamp: time.Now()},
        {DeviceID: "sensor3", Temperature: 25.0, Humidity: 15.0, Timestamp: time.Now()},
    }

    for _, reading := range readings {
        alert := AlertContext{
            Sensor:   &reading,
            Alert:    false,
            Severity: "",
            Message:  "",
        }

        // Execute rules in priority order
        grule.Execute(context.Background(), "CriticalTemperature", &alert)
        if !alert.Alert {
            grule.Execute(context.Background(), "TemperatureAlert", &alert)
        }
        grule.Execute(context.Background(), "HumidityAlert", &alert)

        if alert.Alert {
            fmt.Printf("ALERT [%s] Device %s: %s\n",
                alert.Severity, reading.DeviceID, alert.Message)
        } else {
            fmt.Printf("OK Device %s: Normal conditions\n", reading.DeviceID)
        }
    }
}
```

## Configuration Examples

### High-Performance Configuration

```go
cfg := engine.Config{
    Type:            engine.ARC,              // Adaptive cache
    Size:            10000,                   // Large cache
    CleanupInterval: 30,                      // Frequent cleanup
    TTL:             3600,                    // 1 hour TTL
    Partition:       runtime.NumCPU(),        // Match CPU cores
    FactName:        "BusinessFact",
}
```

### Memory-Constrained Configuration

```go
cfg := engine.Config{
    Type:            engine.LRU,              // Simple cache
    Size:            500,                     // Small cache
    CleanupInterval: 300,                     // Less frequent cleanup
    TTL:             1800,                    // 30 minutes TTL
    Partition:       2,                       // Few partitions
    FactName:        "Fact",
}
```

### Development Configuration

```go
cfg := engine.Config{
    Type:            engine.LRU,
    Size:            100,                     // Small for development
    CleanupInterval: 10,                      // Fast cleanup
    TTL:             300,                     // Short TTL
    Partition:       1,                       // Single partition
    FactName:        "DevFact",
}
```

## Error Handling

### Proper Error Handling

```go
grule := engine.NewPartitionEngine(cfg, nil)
defer grule.Close()

rule := "MyRule"
statement := `rule MyRule "Example rule" salience 10 {
    when
        Fact.Value > 10
    then
        Fact.Result = true;
        Retract("MyRule");
}`

// Add rule with error handling
if err := grule.AddRule(rule, statement, 300); err != nil {
    log.Printf("Failed to add rule: %v", err)
    return
}

// Execute with error handling
fact := MyFact{Value: 15}
if err := grule.Execute(context.Background(), rule, &fact); err != nil {
    log.Printf("Rule execution failed: %v", err)
    return
}
```

### Checking Rule Existence

```go
if grule.ContainsRule("MyRule") {
    fmt.Println("Rule exists")
} else {
    fmt.Println("Rule not found")
}
```

## Testing Examples

### Unit Test Example

```go
func TestDiscountRule(t *testing.T) {
    cfg := engine.Config{
        Type:            engine.LRU,
        Size:            100,
        CleanupInterval: 10,
        TTL:             60,
        Partition:       1,
        FactName:        "DiscountFact",
    }

    grule := engine.NewPartitionEngine(cfg, nil)
    defer grule.Close()

    rule := `rule TestDiscount "Test discount rule" salience 10 {
        when
            DiscountFact.Amount > 100
        then
            DiscountFact.Discount = 15;
            Retract("TestDiscount");
    }`

    err := grule.AddRule("TestDiscount", rule, 60)
    require.NoError(t, err)

    fact := struct {
        Amount   int
        Discount int
    }{Amount: 150}

    err = grule.Execute(context.Background(), "TestDiscount", &fact)
    require.NoError(t, err)
    assert.Equal(t, 15, fact.Discount)
}
```

## Integration Examples

### HTTP Server with Rule Engine

```go
package main

import (
    "encoding/json"
    "net/http"
    "github.com/hungpdn/grule-plus/engine"
)

type RuleEngine struct {
    engine *engine.PartitionEngine
}

func NewRuleEngine() *RuleEngine {
    cfg := engine.Config{
        Type:            engine.LRU,
        Size:            1000,
        CleanupInterval: 60,
        TTL:             300,
        Partition:       4,
        FactName:        "RequestFact",
    }

    return &RuleEngine{
        engine: engine.NewPartitionEngine(cfg, nil),
    }
}

func (re *RuleEngine) EvaluateRules(w http.ResponseWriter, r *http.Request) {
    var fact map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&fact); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Apply business rules
    rule := "BusinessRule"
    // ... rule logic ...

    err := re.engine.Execute(r.Context(), rule, &fact)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(fact)
}

func main() {
    re := NewRuleEngine()
    defer re.engine.Close()

    http.HandleFunc("/evaluate", re.EvaluateRules)
    http.ListenAndServe(":8080", nil)
}
```
