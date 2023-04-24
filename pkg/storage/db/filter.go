package db

import (
	"github.com/krixlion/dev_forum-lib/filter"
	"go.mongodb.org/mongo-driver/bson"
)

func filterToBSON(params []filter.Parameter) (bson.D, error) {
	filterDoc := make(bson.D, 0, len(params))

	for _, param := range params {
		mongoOperator, err := toMongoOperator(param.Operator)
		if err != nil {
			return nil, err
		}

		element := bson.E{
			Key:   param.Attribute,
			Value: bson.D{{Key: mongoOperator, Value: param.Value}},
		}

		filterDoc = append(filterDoc, element)
	}

	return filterDoc, nil
}

func toMongoOperator(op filter.Operator) (string, error) {
	switch op {
	case filter.Equal:
		return "$eq", nil
	case filter.NotEqual:
		return "$ne", nil
	case filter.GreaterThan:
		return "$gt", nil
	case filter.GreaterThanOrEqual:
		return "$gte", nil
	case filter.LesserThan:
		return "$lt", nil
	case filter.LesserThanOrEqual:
		return "$lte", nil
	default:
		return "", filter.ErrInvalidOperator
	}
}
