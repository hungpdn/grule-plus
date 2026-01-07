package discountpb

import (
	"testing"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
)

func TestProtobufDiscountRule(t *testing.T) {
	// Create a rule that applies discount based on customer type and amount
	rule := `
rule DiscountRule "Apply discount based on customer type and amount" salience 10 {
	when
		DiscountFact.Amount > 100 && DiscountFact.CustomerType == "premium"
	then
		DiscountFact.Discount = 20;
		Retract("DiscountRule");
}

rule LoyaltyRule "Apply loyalty discount for high value orders" salience 5 {
	when
		DiscountFact.Amount > 500
	then
		DiscountFact.Discount = DiscountFact.Discount + 10;
		Retract("LoyaltyRule");
}
`

	// Create protobuf fact
	discountFact := &DiscountFact{
		Amount:       600,
		CustomerType: "premium",
		Discount:     0,
		Tags:         []string{},
	}

	// Create data context and add the protobuf fact
	dataContext := ast.NewDataContext()
	err := dataContext.Add("DiscountFact", discountFact)
	if err != nil {
		t.Fatalf("Failed to add protobuf fact to data context: %v", err)
	}

	// Build the knowledge base
	lib := ast.NewKnowledgeLibrary()
	ruleBuilder := builder.NewRuleBuilder(lib)
	err = ruleBuilder.BuildRuleFromResource("Test", "0.1.1", pkg.NewBytesResource([]byte(rule)))
	if err != nil {
		t.Fatalf("Failed to build rule: %v", err)
	}

	// Create engine and execute
	knowledgeBase, err := lib.NewKnowledgeBaseInstance("Test", "0.1.1")
	if err != nil {
		t.Fatalf("Failed to create knowledge base: %v", err)
	}

	eng := engine.NewGruleEngine()
	err = eng.Execute(dataContext, knowledgeBase)
	if err != nil {
		t.Fatalf("Failed to execute rules: %v", err)
	}

	// Verify results
	expectedDiscount := int32(30) // 20 from premium + 10 from loyalty
	if discountFact.Discount != expectedDiscount {
		t.Errorf("Expected discount %d, got %d", expectedDiscount, discountFact.Discount)
	}
}

func TestProtobufOrderRule(t *testing.T) {
	// Create a rule that processes order items
	orderRule := `
rule ProcessOrder "Process order and apply discount based on total" salience 10 {
	when
		OrderFact.TotalAmount > 200
	then
		OrderFact.Discount.Amount = 50;
		OrderFact.Discount.CustomerType = "bulk_order";
		Retract("ProcessOrder");
}

rule ApplyItemDiscount "Apply discount to expensive items" salience 5 {
	when
		Item.Price > 100
	then
		Item.Quantity = Item.Quantity + 1; // Give bonus item for expensive purchases
		Retract("ApplyItemDiscount");
}
`

	// Create protobuf facts
	orderFact := &OrderFact{
		OrderId:     "order-123",
		TotalAmount: 250,
		CustomerId:  "customer-456",
		Items: []*Item{
			{
				ProductId: "item-456",
				Name:      "Premium Widget",
				Price:     150,
				Quantity:  1,
				Category:  "electronics",
			},
		},
		Discount: &DiscountFact{
			Amount:       0,
			CustomerType: "",
			Discount:     0,
			Tags:         []string{},
		},
	}

	// Create data context and add facts
	dataContext := ast.NewDataContext()
	err := dataContext.Add("OrderFact", orderFact)
	if err != nil {
		t.Fatalf("Failed to add order fact: %v", err)
	}

	// Add the item as a separate fact for the rule engine
	item := orderFact.Items[0]
	err = dataContext.Add("Item", item)
	if err != nil {
		t.Fatalf("Failed to add item fact: %v", err)
	}

	// Build knowledge base
	lib := ast.NewKnowledgeLibrary()
	ruleBuilder := builder.NewRuleBuilder(lib)
	err = ruleBuilder.BuildRuleFromResource("OrderTest", "0.1.1", pkg.NewBytesResource([]byte(orderRule)))
	if err != nil {
		t.Fatalf("Failed to build order rule: %v", err)
	}

	// Execute rules
	knowledgeBase, err := lib.NewKnowledgeBaseInstance("OrderTest", "0.1.1")
	if err != nil {
		t.Fatalf("Failed to create knowledge base: %v", err)
	}

	eng := engine.NewGruleEngine()
	err = eng.Execute(dataContext, knowledgeBase)
	if err != nil {
		t.Fatalf("Failed to execute order rules: %v", err)
	}

	// Verify order discount
	if orderFact.Discount.Amount != 50 {
		t.Errorf("Expected discount amount 50, got %d", orderFact.Discount.Amount)
	}
	if orderFact.Discount.CustomerType != "bulk_order" {
		t.Errorf("Expected customer type 'bulk_order', got '%s'", orderFact.Discount.CustomerType)
	}

	// Verify item bonus
	if item.Quantity != 2 {
		t.Errorf("Expected quantity 2 (bonus item), got %d", item.Quantity)
	}
}
