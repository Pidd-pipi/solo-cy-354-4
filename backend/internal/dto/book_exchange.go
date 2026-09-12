package dto

// CreateBookExchangeRequest is the payload for posting a book swap.
type CreateBookExchangeRequest struct {
	OfferBook   string `json:"offer_book" binding:"required,min=1,max=64"`
	WantBook    string `json:"want_book" binding:"required,min=1,max=64"`
	Description string `json:"description"`
}
