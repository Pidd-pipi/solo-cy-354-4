// Package tradestate holds the trade order lifecycle rules: which actor role
// may move an order from one status to another and which timestamps the
// transition records. It performs no I/O; callers persist whatever the rules
// approve. Every order action (create, buyer confirm, seller confirm,
// cancel) must go through Initial/Guard so there is exactly one place that
// defines the state machine.
package tradestate

import (
	"errors"
	"time"

	"github.com/lp/campus-market/internal/constants"
	"github.com/lp/campus-market/internal/model"
)

// Actor role for an order action.
type Role int

const (
	RoleUnknown Role = iota
	// RoleBuyer only the order buyer may perform the action.
	RoleBuyer
	// RoleSeller only the order seller may perform the action.
	RoleSeller
	// RoleParticipant either the buyer or the seller may perform the action.
	RoleParticipant
)

// Actions driving the order lifecycle.
const (
	ActionBuyerConfirm  = "buyer_confirm"
	ActionSellerConfirm = "seller_confirm"
	ActionCancel        = "cancel"
)

// Sentinel errors returned by Guard; the service layer maps them onto the
// same 403/409 responses every action returns.
var (
	ErrIllegalRole       = errors.New("tradestate: actor role not allowed")
	ErrIllegalTransition = errors.New("tradestate: status transition not allowed")
)

// Rule describes one allowed transition and the timestamps it records.
// The boolean flags select which confirmation columns the transition
// stamps; columns left false stay untouched.
type Rule struct {
	Action            string
	Role              Role
	From              string
	To                string
	StampBuyer        bool
	StampSeller       bool
	StampCompleted    bool
	ProductOnComplete string // product status to set when To is reached, "" means no change
}

// rules is the single state-machine table shared by every order action.
//
//	pending --buyer confirm-->  confirmed
//	confirmed --seller confirm--> completed (product -> sold)
//	pending --buyer/seller cancel--> cancelled
var rules = map[string][]Rule{
	ActionBuyerConfirm: {
		{
			Action: ActionBuyerConfirm, Role: RoleBuyer,
			From: constants.TradeStatusPending, To: constants.TradeStatusConfirmed,
			StampBuyer: true,
		},
	},
	ActionSellerConfirm: {
		{
			Action: ActionSellerConfirm, Role: RoleSeller,
			From: constants.TradeStatusConfirmed, To: constants.TradeStatusCompleted,
			StampSeller: true, StampCompleted: true,
			ProductOnComplete: constants.ProductStatusSold,
		},
	},
	ActionCancel: {
		{
			Action: ActionCancel, Role: RoleParticipant,
			From: constants.TradeStatusPending, To: constants.TradeStatusCancelled,
		},
	},
}

// Initial returns the status assigned when an order is created.
func Initial() string { return constants.TradeStatusPending }

// ActiveStatuses lists order statuses that block placing another order for
// the same product (pending/confirmed; not finished, not cancelled).
func ActiveStatuses() []string {
	return []string{constants.TradeStatusPending, constants.TradeStatusConfirmed}
}

// ColumnUpdates maps the rule timestamp flags onto the columns that must be
// stamped with now. Column names live here because they are part of the
// transition rule; the repository merely persists the returned map merged
// with the new status.
func (r *Rule) ColumnUpdates(now time.Time) map[string]interface{} {
	updates := map[string]interface{}{}
	if r.StampBuyer {
		updates["buyer_confirmed_at"] = now
	}
	if r.StampSeller {
		updates["seller_confirmed_at"] = now
	}
	if r.StampCompleted {
		updates["completed_at"] = now
	}
	return updates
}

// Guard is the single gate every order action passes. It first verifies the
// acting user matches the role the transition requires, then verifies the
// order's current status allows the action. It returns the matched rule for
// the caller to persist.
func Guard(order *model.TradeOrder, action string, userID uint) (*Rule, error) {
	candidates, ok := rules[action]
	if !ok {
		return nil, ErrIllegalTransition
	}
	for i := range candidates {
		rule := &candidates[i]
		if !roleAllowed(rule.Role, order, userID) {
			continue
		}
		if order.Status != rule.From {
			return nil, ErrIllegalTransition
		}
		return rule, nil
	}
	// No candidate matched the role: that is an authorization problem rather
	// than a state problem, regardless of the order's current status.
	return nil, ErrIllegalRole
}

func roleAllowed(role Role, order *model.TradeOrder, userID uint) bool {
	switch role {
	case RoleBuyer:
		return order.BuyerID == userID
	case RoleSeller:
		return order.SellerID == userID
	case RoleParticipant:
		return order.BuyerID == userID || order.SellerID == userID
	default:
		return false
	}
}
