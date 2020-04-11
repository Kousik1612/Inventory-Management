package main

//Products ...
type Products struct {
	ProductType      string        `json:"product_type,omitempty"`
	SellerID         string        `json:"seller_id,omitempty"`
	Brand            string        `json:"brand,omitempty"`
	Size             string        `json:"size,omitempty"`
	Metadata         []interface{} `json:"metadata,omitempty"`
	Location         string        `json:"location,omitempty"`
	LocationID       string        `json:"location_id,omitempty"`
	IsAvailable      bool          `json:"isavailable,omitempty"`
	MetadataResponse interface{}   `json:"metadataresult,omitempty"`
	ProductIDS       []string      `json:"product_ids,omitempty"`
}

//Order ...
type Order struct {
	OrderResult    string `json:"message,omitempty"`
	InvoiceMessage string `json:"invoice,omitempty"`
}
