package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/qri-io/jsonschema"
)

func main() {
	ctx := context.Background()
	var schemaData = []byte(`{
    "$schema": "http://json-schema.org/draft-07/schema",
    "$id": "http://example.com/example.json",
    "type": "object",
    "title": "The root schema",
    "description": "The root schema comprises the entire JSON document.",
    "default": {},
    "examples": [
        {
            "@context": "https://schema.org",
            "@type": "VisualArtwork",
            "name": "La trahison des images",
            "alternateName": "The Treachery of Images",
            "image": "http://upload.wikimedia.org/wikipedia/en/b/b9/MagrittePipe.jpg",
            "description": "The painting shows a pipe. Below it, Magritte...",
            "creator": [
                {
                    "@type": "Person",
                    "name": "René Magritte",
                    "sameAs": "https://www.freebase.com/m/06h88"
                }
            ],
            "width": [
                {
                    "@type": "Distance",
                    "name": "940 mm"
                }
            ],
            "height": [
                {
                    "@type": "Distance",
                    "name": "635 mm"
                }
            ],
            "artMedium": "oil",
            "artworkSurface": "canvas"
        }
    ],
    "required": [
        "@context",
        "@type",
        "name",
        "alternateName",
        "image",
        "description",
        "creator",
        "width",
        "height",
        "artMedium",
        "artworkSurface"
    ],
    "properties": {
        "@context": {
            "$id": "#/properties/%40context",
            "type": "string",
            "title": "The @context schema",
            "description": "An explanation about the purpose of this instance.",
            "default": "",
            "examples": [
                "https://schema.org"
            ]
        },
        "@type": {
            "$id": "#/properties/%40type",
            "type": "string",
            "title": "The @type schema",
            "description": "An explanation about the purpose of this instance.",
            "default": "",
            "examples": [
                "VisualArtwork"
            ]
        },
        "name": {
            "$id": "#/properties/name",
            "type": "string",
            "title": "The name schema",
            "description": "An explanation about the purpose of this instance.",
            "default": "",
            "examples": [
                "La trahison des images"
            ]
        },
        "alternateName": {
            "$id": "#/properties/alternateName",
            "type": "string",
            "title": "The alternateName schema",
            "description": "An explanation about the purpose of this instance.",
            "default": "",
            "examples": [
                "The Treachery of Images"
            ]
        },
        "image": {
            "$id": "#/properties/image",
            "type": "string",
            "title": "The image schema",
            "description": "An explanation about the purpose of this instance.",
            "default": "",
            "examples": [
                "http://upload.wikimedia.org/wikipedia/en/b/b9/MagrittePipe.jpg"
            ]
        },
        "description": {
            "$id": "#/properties/description",
            "type": "string",
            "title": "The description schema",
            "description": "An explanation about the purpose of this instance.",
            "default": "",
            "examples": [
                "The painting shows a pipe. Below it, Magritte..."
            ]
        },
        "creator": {
            "$id": "#/properties/creator",
            "type": "array",
            "title": "The creator schema",
            "description": "An explanation about the purpose of this instance.",
            "default": [],
            "examples": [
                [
                    {
                        "@type": "Person",
                        "name": "René Magritte",
                        "sameAs": "https://www.freebase.com/m/06h88"
                    }
                ]
            ],
            "additionalItems": true,
            "items": {
                "$id": "#/properties/creator/items",
                "anyOf": [
                    {
                        "$id": "#/properties/creator/items/anyOf/0",
                        "type": "object",
                        "title": "The first anyOf schema",
                        "description": "An explanation about the purpose of this instance.",
                        "default": {},
                        "examples": [
                            {
                                "@type": "Person",
                                "name": "René Magritte",
                                "sameAs": "https://www.freebase.com/m/06h88"
                            }
                        ],
                        "required": [
                            "@type",
                            "name",
                            "sameAs"
                        ],
                        "properties": {
                            "@type": {
                                "$id": "#/properties/creator/items/anyOf/0/properties/%40type",
                                "type": "string",
                                "title": "The @type schema",
                                "description": "An explanation about the purpose of this instance.",
                                "default": "",
                                "examples": [
                                    "Person"
                                ]
                            },
                            "name": {
                                "$id": "#/properties/creator/items/anyOf/0/properties/name",
                                "type": "string",
                                "title": "The name schema",
                                "description": "An explanation about the purpose of this instance.",
                                "default": "",
                                "examples": [
                                    "René Magritte"
                                ]
                            },
                            "sameAs": {
                                "$id": "#/properties/creator/items/anyOf/0/properties/sameAs",
                                "type": "string",
                                "title": "The sameAs schema",
                                "description": "An explanation about the purpose of this instance.",
                                "default": "",
                                "examples": [
                                    "https://www.freebase.com/m/06h88"
                                ]
                            }
                        },
                        "additionalProperties": true
                    }
                ]
            }
        },
        "width": {
            "$id": "#/properties/width",
            "type": "array",
            "title": "The width schema",
            "description": "An explanation about the purpose of this instance.",
            "default": [],
            "examples": [
                [
                    {
                        "@type": "Distance",
                        "name": "940 mm"
                    }
                ]
            ],
            "additionalItems": true,
            "items": {
                "$id": "#/properties/width/items",
                "anyOf": [
                    {
                        "$id": "#/properties/width/items/anyOf/0",
                        "type": "object",
                        "title": "The first anyOf schema",
                        "description": "An explanation about the purpose of this instance.",
                        "default": {},
                        "examples": [
                            {
                                "@type": "Distance",
                                "name": "940 mm"
                            }
                        ],
                        "required": [
                            "@type",
                            "name"
                        ],
                        "properties": {
                            "@type": {
                                "$id": "#/properties/width/items/anyOf/0/properties/%40type",
                                "type": "string",
                                "title": "The @type schema",
                                "description": "An explanation about the purpose of this instance.",
                                "default": "",
                                "examples": [
                                    "Distance"
                                ]
                            },
                            "name": {
                                "$id": "#/properties/width/items/anyOf/0/properties/name",
                                "type": "string",
                                "title": "The name schema",
                                "description": "An explanation about the purpose of this instance.",
                                "default": "",
                                "examples": [
                                    "940 mm"
                                ]
                            }
                        },
                        "additionalProperties": true
                    }
                ]
            }
        },
        "height": {
            "$id": "#/properties/height",
            "type": "array",
            "title": "The height schema",
            "description": "An explanation about the purpose of this instance.",
            "default": [],
            "examples": [
                [
                    {
                        "@type": "Distance",
                        "name": "635 mm"
                    }
                ]
            ],
            "additionalItems": true,
            "items": {
                "$id": "#/properties/height/items",
                "anyOf": [
                    {
                        "$id": "#/properties/height/items/anyOf/0",
                        "type": "object",
                        "title": "The first anyOf schema",
                        "description": "An explanation about the purpose of this instance.",
                        "default": {},
                        "examples": [
                            {
                                "@type": "Distance",
                                "name": "635 mm"
                            }
                        ],
                        "required": [
                            "@type",
                            "name"
                        ],
                        "properties": {
                            "@type": {
                                "$id": "#/properties/height/items/anyOf/0/properties/%40type",
                                "type": "string",
                                "title": "The @type schema",
                                "description": "An explanation about the purpose of this instance.",
                                "default": "",
                                "examples": [
                                    "Distance"
                                ]
                            },
                            "name": {
                                "$id": "#/properties/height/items/anyOf/0/properties/name",
                                "type": "string",
                                "title": "The name schema",
                                "description": "An explanation about the purpose of this instance.",
                                "default": "",
                                "examples": [
                                    "635 mm"
                                ]
                            }
                        },
                        "additionalProperties": true
                    }
                ]
            }
        },
        "artMedium": {
            "$id": "#/properties/artMedium",
            "type": "string",
            "title": "The artMedium schema",
            "description": "An explanation about the purpose of this instance.",
            "default": "",
            "examples": [
                "oil"
            ]
        },
        "artworkSurface": {
            "$id": "#/properties/artworkSurface",
            "type": "string",
            "title": "The artworkSurface schema",
            "description": "An explanation about the purpose of this instance.",
            "default": "",
            "examples": [
                "canvas"
            ]
        }
    },
    "additionalProperties": true
}`)

	rs := &jsonschema.Schema{}
	if err := json.Unmarshal(schemaData, rs); err != nil {
		panic("unmarshal schema: " + err.Error())
	}

	var valid = []byte(`{
      "@context": "https://schema.org",
      "@type": "VisualArtwork",
      "name": "La trahison des images",
      "alternateName": "The Treachery of Images",
      "image": "http://upload.wikimedia.org/wikipedia/en/b/b9/MagrittePipe.jpg",
      "description": "The painting shows a pipe. Below it, Magritte...",
      "creator": [
        {
          "@type": "Person",
          "name": "René Magritte",
          "sameAs": "https://www.freebase.com/m/06h88"
        }
      ],
      "width": [
        {
          "@type": "Distance",
          "name": "940 mm"
        }
      ],
      "height": [
        {
          "@type": "Distance",
          "name": "635 mm"
        }
      ],
      "artMedium": "oil",
      "artworkSurface": "canvas"
    }`)
	errs, err := rs.ValidateBytes(ctx, valid)
	if err != nil {
		panic(err)
	} else {
		fmt.Println("ok, no panic")
	}

	if len(errs) > 0 {
		fmt.Println(errs[0].Error())
	} else {
		fmt.Println("ok, no error")
	}

}
