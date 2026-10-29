package service

import "testing"

func TestDeliveryDetailsInvalid(t *testing.T) {
	svc := &orderService{}
	valid := DeliveryInfoInput{
		RecipientName: "Alex Smith",
		Phone:         "01131234567",
		Address:       "10 High Street",
		City:          "London",
		State:         "Greater London",
	}
	tests := []struct {
		name  string
		input DeliveryInfoInput
		field string
	}{
		{name: "missing name", input: withoutDeliveryField(valid, "recipient_name"), field: "recipient_name"},
		{name: "missing phone", input: withoutDeliveryField(valid, "phone"), field: "phone"},
		{name: "missing address", input: withoutDeliveryField(valid, "address"), field: "address"},
		{name: "missing city", input: withoutDeliveryField(valid, "city"), field: "city"},
		{name: "missing state", input: withoutDeliveryField(valid, "state"), field: "state"},
		{
			name: "spaces only",
			input: DeliveryInfoInput{
				RecipientName: "   ",
				Phone:         "01131234567",
				Address:       "10 High Street",
				City:          "London",
				State:         "Greater London",
			},
			field: "recipient_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.input
			errs := svc.ValidateDeliveryInfo(&input)
			if errs[tt.field] == "" {
				t.Errorf("expected %s to be invalid, got %v", tt.field, errs)
			}
		})
	}
}

func withoutDeliveryField(input DeliveryInfoInput, field string) DeliveryInfoInput {
	switch field {
	case "recipient_name":
		input.RecipientName = ""
	case "phone":
		input.Phone = ""
	case "address":
		input.Address = ""
	case "city":
		input.City = ""
	case "state":
		input.State = ""
	}
	return input
}
