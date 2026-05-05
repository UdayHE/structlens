package model_test

import (
	"fmt"

	"structlens/internal/model"
)

func ExampleNode() {
	jsonRoot := &model.Node{
		Name: "root",
		Type: model.NodeTypeObject,
		Path: "$",
		Children: []*model.Node{
			{
				Name: "users",
				Type: model.NodeTypeArray,
				Path: "$.users",
			},
		},
	}

	xmlRoot := &model.Node{
		Name: "book",
		Type: model.NodeTypeObject,
		Path: "/book",
		Attributes: map[string]string{
			"id": "bk-1001",
		},
		Children: []*model.Node{
			{
				Name: "title",
				Type: model.NodeTypeString,
				Path: "/book/title",
			},
		},
	}

	fmt.Println(jsonRoot.Name, jsonRoot.Children[0].Path)
	fmt.Println(xmlRoot.Name, xmlRoot.Attributes["id"])
	// Output:
	// root $.users
	// book bk-1001
}
