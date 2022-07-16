// Code generated by goctl. DO NOT EDIT!
package model

import (
    "context"
    "time"

    {{if .Cache}}"github.com/zeromicro/go-zero/core/stores/monc"{{else}}"github.com/zeromicro/go-zero/core/stores/mon"{{end}}
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

{{if .Cache}}var prefix{{.Type}}CacheKey = "cache:{{.lowerType}}:"{{end}}

type {{.lowerType}}Model interface{
    Insert(ctx context.Context,data *{{.Type}}) error
    FindOne(ctx context.Context,id string) (*{{.Type}}, error)
    Update(ctx context.Context,data *{{.Type}}) error
    Delete(ctx context.Context,id string) error
}

type default{{.Type}}Model struct {
    conn {{if .Cache}}*monc.Model{{else}}*mon.Model{{end}}
}

func newDefault{{.Type}}Model(conn {{if .Cache}}*monc.Model{{else}}*mon.Model{{end}}) *default{{.Type}}Model {
    return &default{{.Type}}Model{conn: conn}
}


func (m *default{{.Type}}Model) Insert(ctx context.Context, data *{{.Type}}) error {
    if !data.ID.IsZero() {
        data.ID = primitive.NewObjectID()
        data.CreateAt = time.Now()
        data.UpdateAt = time.Now()
    }

    {{if .Cache}}key := prefix{{.Type}}CacheKey + data.ID.Hex(){{end}}
    _, err := m.conn.InsertOne(ctx, {{if .Cache}}key, {{end}} data)
    return err
}

func (m *default{{.Type}}Model) FindOne(ctx context.Context, id string) (*{{.Type}}, error) {
    oid, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return nil, ErrInvalidObjectId
    }

    var data {{.Type}}
    {{if .Cache}}key := prefix{{.Type}}CacheKey + data.ID.Hex(){{end}}
    err = m.conn.FindOne(ctx, {{if .Cache}}key, {{end}}&data, bson.M{"_id": oid})
    switch err {
    case nil:
        return &data, nil
    case {{if .Cache}}monc{{else}}mon{{end}}.ErrNotFound:
        return nil, ErrNotFound
    default:
        return nil, err
    }
}

func (m *default{{.Type}}Model) Update(ctx context.Context, data *{{.Type}}) error {
    data.UpdateAt = time.Now()
    {{if .Cache}}key := prefix{{.Type}}CacheKey + data.ID.Hex(){{end}}
    _, err := m.conn.ReplaceOne(ctx, {{if .Cache}}key, {{end}}bson.M{"_id": data.ID}, data)
    return err
}

func (m *default{{.Type}}Model) Delete(ctx context.Context, id string) error {
    oid, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return ErrInvalidObjectId
    }
	{{if .Cache}}key := prefix{{.Type}}CacheKey +id{{end}}
    _, err = m.conn.DeleteOne(ctx, {{if .Cache}}key, {{end}}bson.M{"_id": oid})
	return err
}
