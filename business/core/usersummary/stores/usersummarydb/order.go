package usersummarydb

import (
	"fmt"

	"github.com/sco1237896/sco-backend/business/core/usersummary"
	"github.com/sco1237896/sco-backend/business/data/order"
)

var orderByFields = map[string]string{
	usersummary.OrderByUserID:   "user_id",
	usersummary.OrderByUserName: "user_name",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
