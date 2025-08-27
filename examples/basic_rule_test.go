package examples

import (
	"context"
	"testing"

	"github.com/hungpdn/grule-plus/engine"
	"github.com/stretchr/testify/assert"
)

func TestBasicRule(t *testing.T) {
	cfg := engine.Config{
		Type:            0, // LRU
		Size:            1000,
		CleanupInterval: 10,
		TTL:             60,
		Partition:       4,
	}
	grule := engine.NewPartitionEngine(cfg, nil)

	rule := "DiscountRule"
	statement := `rule DiscountRule "Apply discount" salience 10 {
	 			when 
					DiscountFact.Amount > 100 
				then 
					DiscountFact.Discount = 10; 
					Retract("DiscountRule");
				}`

	_ = grule.AddRule(rule, statement, 60)

	fact := struct {
		Amount   int
		Discount int
	}{Amount: 150}

	err := grule.Execute(context.Background(), rule, &fact)

	assert.NoError(t, err)
	assert.Equal(t, 10, fact.Discount)
}
