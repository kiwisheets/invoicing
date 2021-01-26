package logger

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/sirupsen/logrus"
)

type contextKeyType struct{}

var (
	loggerContextKey = contextKeyType(struct{}{})
)

type LogrusExtension struct {
	Logger *logrus.Entry
}

var _ interface {
	graphql.HandlerExtension
	graphql.FieldInterceptor
} = LogrusExtension{}

func (n LogrusExtension) ExtensionName() string {
	return "LogrusExtension"
}

func (n LogrusExtension) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

func (n LogrusExtension) InterceptField(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	oc := graphql.GetOperationContext(ctx)
	fc := graphql.GetFieldContext(ctx)

	// TODO: get request ID from request context (gin middleware?)
	ctxLogger := n.Logger.WithFields(logrus.Fields{
		"operation": oc.OperationName,
		"field":     fc.Field.Name,
	})
	ctx = new(ctx, ctxLogger)
	return next(ctx)
}

func new(ctx context.Context, ctxLogger *logrus.Entry) context.Context {
	return context.WithValue(ctx, loggerContextKey, ctxLogger)
}

func From(ctx context.Context) *logrus.Entry {
	l, _ := ctx.Value(loggerContextKey).(*logrus.Entry)
	return l
}
