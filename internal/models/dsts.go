// Package models implements the base structures and functions for managing data models in the application.
package models

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/goccy/go-json"
	ci "github.com/kubex-ecosystem/domus/internal/interfaces"
	"github.com/kubex-ecosystem/logz"
)

type Model interface {
	ci.IReference // Aqui tem nome e ID (uuid)
	Validate() error
}

var ModelList = make([]interface{}, 0)
var ModelRegistryMap = map[reflect.Type]struct{}{}

type ModelRegistryImpl[T Model] struct {
	Dt T      `json:"data"`
	St []byte `json:"status"`
}

type ModelRegistryInterface interface {
	GetType() reflect.Type
	FromModel(model interface{}) ModelRegistryInterface
	FromSerialized(data []byte) (ModelRegistryInterface, error)
	ToModel() interface{}
}

func (m *ModelRegistryImpl[T]) GetType() reflect.Type { return reflect.TypeFor[T]() }

func (m *ModelRegistryImpl[T]) FromModel(model interface{}) ModelRegistryInterface {
	if model == nil {
		return nil
	}
	instance, ok := model.(T)
	if !ok {
		return nil
	}
	val := reflect.ValueOf(instance)
	if !val.IsValid() || (val.Kind() == reflect.Ptr && val.IsNil()) {
		return nil
	}
	if err := instance.Validate(); err != nil {
		return nil
	}
	if _, ok := ModelRegistryMap[reflect.TypeOf(instance)]; !ok {
		return nil
	}
	m.Dt = instance
	m.St, _ = json.Marshal(instance)
	return m
}

func (m *ModelRegistryImpl[T]) FromSerialized(data []byte) (ModelRegistryInterface, error) {
	var mdr ModelRegistryImpl[T]
	if err := json.Unmarshal(data, &mdr); err != nil {
		return nil, err
	}
	// Retorna o tipo que está implícito na estrutura pelo generic T
	// Assim não é preciso armazenar o tipo do modelo
	// Verifica se o tipo do modelo está registrado
	typ := reflect.TypeOf(mdr.Dt)
	if typ == nil {
		return nil, fmt.Errorf("model %s not found", mdr.GetType())
	}
	if _, ok := ModelRegistryMap[typ]; !ok {
		return nil, fmt.Errorf("model %s not found", mdr.GetType())
	}
	return &mdr, nil
}

func (m *ModelRegistryImpl[T]) ToModel() interface{} {
	if any(m.Dt) == nil {
		return nil
	}
	return m.Dt
}

func RegisterModel(modelType reflect.Type) error {
	// Ferrou porque não tem mais como guardar o nome.. rsrs
	if _, exists := ModelRegistryMap[modelType]; exists {
		return fmt.Errorf("model %s já registrado", modelType.String())
	}
	// O map armazena valores pelo tipo do modelo, então como estamos só
	// registrando o tipo, não precisamos guardar valor. O nome está implícito
	// na interface Model. Só implementar lá. rsrs
	ModelRegistryMap[modelType] = struct{}{}
	return nil
}

func NewModelRegistry[T Model]() ModelRegistryInterface {
	return &ModelRegistryImpl[T]{}
}

func NewModelRegistryFromModel[T Model](model interface{}) ModelRegistryInterface {
	mr := ModelRegistryImpl[T]{}
	return mr.FromModel(model)
}

func NewModelRegistryFromSerialized[T Model](data []byte) (ModelRegistryInterface, error) {
	mr := ModelRegistryImpl[T]{}
	return mr.FromSerialized(data)
}

func ParseConditionClause(conditions ...any) (string, []any, error) {
	if len(conditions) == 0 {
		return "", nil, fmt.Errorf("no conditions provided")
	}

	var whereClause strings.Builder
	var clauseSliceObj = make([]any, 0) // É slice de interfaces, então a merda que vier de valor é só jogar e

	// Cada posição da condição pode ser um slice, um map ou qualquer outra coisa...
	for _, condition := range conditions {
		value := reflect.ValueOf(condition)

		// Então primeiro detecto se é slice, se for, trato como slice
		switch reflect.TypeOf(condition).Kind() {
		case reflect.Slice, reflect.SliceOf(reflect.TypeFor[map[string]string]()).Kind():
			if value.Len() == 0 {
				// If the slice is empty, assign the deserialized object to the slice
				logz.Log("debug", "Query conditions are empty")
				continue
			} else if value.Len() == 1 {
				// If the slice has only one element, assign it directly
				clauseSliceObj = append(clauseSliceObj, value.Index(0).Interface())
				continue
			} else if value.Len() > 1 {
				if value.Type().Elem().Kind() == reflect.Map {
					for i := 0; i < value.Len(); i++ {
						clauseSliceObj = append(clauseSliceObj, value.Index(i).Interface())
					}
				}
				continue
			} else {
				for i := 0; i < value.Len(); i++ {
					clauseSliceObj = append(clauseSliceObj, value.Index(i).Interface())
				}
			}
		// Segundo verifico se é map, se for, trato como map
		case reflect.Map:
			if value.Len() == 0 {
				// If the map is empty, assign the deserialized object to the map
				clauseSliceObj = reflect.ValueOf(conditions).Interface().([]any)
			} else {
				for _, key := range value.MapKeys() {
					clauseSliceObj = append(clauseSliceObj, fmt.Sprintf("%s = ?", key.String()))
					clauseSliceObj = append(clauseSliceObj, value.MapIndex(key).Interface())
				}
			}
		default:
			// If the type is neither a slice nor a map, assign the first object to m.object
			if len(conditions) == 0 {
				logz.Log("debug", "Query conditions are empty")
				continue
			}
			if len(conditions) > 1 {
				logz.Log("debug", "Multiple query conditions found")
				continue
			}
			clauseSliceObj = append(clauseSliceObj, condition)
		}
	}

	// Agora monto a cláusula WHERE
	whereClause.WriteString("")
	for i, condition := range clauseSliceObj {
		switch keyType := condition.(type) {
		case int:
			whereClause.WriteString(fmt.Sprintf("column%d = ?", i))
		case string:
			whereClause.WriteString(fmt.Sprintf("%s = ?", keyType))
		default:
			return "", nil, fmt.Errorf("unsupported type for where clause: %T", keyType)
		}
		if i < len(clauseSliceObj)-1 {
			whereClause.WriteString(" AND ")
		}
	}
	return whereClause.String(), clauseSliceObj, nil
}

func ListRegisteredModels() []string {
	models := make([]string, 0, len(ModelRegistryMap))
	for modelType := range ModelRegistryMap {
		models = append(models, modelType.String())
	}
	return models
}

func GetRegisteredModelTypes() []reflect.Type {
	types := make([]reflect.Type, 0, len(ModelRegistryMap))
	for modelType := range ModelRegistryMap {
		types = append(types, modelType)
	}
	return types
}

func IsModelRegistered(modelType reflect.Type) bool {
	_, exists := ModelRegistryMap[modelType]
	return exists
}

// func GetModelRegistryInstanceByType(modelType reflect.Type) (ModelRegistryInterface, error) {
// 	if !IsModelRegistered(modelType) {
// 		return nil, fmt.Errorf("model %s not registered", modelType.String())
// 	}

// 	// // Cria uma nova instância do ModelRegistryImpl[T] com o tipo correto
// 	// switch modelType {
// 	// case reflect.TypeFor[userstore.User]():
// 	// 	return &ModelRegistryImpl[*userstore.User]{}, nil
// 	// case reflect.TypeOf((*companystore.Company)(nil)).Elem():
// 	// 	return &ModelRegistryImpl[companystore.Company]{
// 	// 		Dt: companystore.Company{},
// 	// 		St: []byte{
// 	// 			0x7b, 0x22, 0x69, 0x64, 0x22, 0x3a, 0x6e, 0x75,
// 	// 			0x6c, 0x6c, 0x2c, 0x22, 0x6e, 0x61, 0x6d,
// 	// 			0x65, 0x22, 0x3a, 0x22, 0x22, 0x2c, 0x22,
// 	// 			0x64, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e,
// 	// 			0x74, 0x22, 0x3a, 0x6e, 0x75, 0x6c, 0x6c,
// 	// 			0x2c, 0x22, 0x63, 0x72, 0x65, 0x61, 0x74,
// 	// 			0x65, 0x64, 0x5f, 0x61, 0x74, 0x22, 0x3a,
// 	// 			0x22, 0x22, 0x2c, 0x22, 0x75, 0x70, 0x64,
// 	// 			0x61, 0x74, 0x65, 0x64, 0x5f, 0x61,
// 	// 			0x74, 0x22, 0x3a, 0x22, 0x22, 0x7d,
// 	// 		},
// 	// 	}, nil
// 	// // Adicione mais casos conforme necessário para outros tipos de modelo
// 	// default:
// 	// 	return nil, fmt.Errorf("no ModelRegistryImpl found for model type %s", modelType.String())
// 	// }
// }

// func GetModelRegistryInstanceByType(modelType reflect.Type) (ModelRegistryInterface, error) {
// 	if !IsModelRegistered(modelType) {
// 		return nil, fmt.Errorf("model %s not registered", modelType.String())
// 	}

// 	// Cria uma nova instância do ModelRegistryImpl[T] com o tipo correto
// 	switch modelType {
// 	case reflect.TypeFor[userstore.User]():
// 		return &ModelRegistryImpl[userstore.User]{}, nil
// 	case reflect.TypeFor[companystore.Company]():
// 		return &ModelRegistryImpl[companystore.Company]{}, nil
// 	// Adicione mais casos conforme necessário para outros tipos de modelo
// 	default:
// 		return nil, fmt.Errorf("no ModelRegistryImpl found for model type %s", modelType.String())
// 	}

// 	return nil, fmt.Errorf("no ModelRegistryImpl found for model type %s", modelType.String())
// }

// func GetModelRegistryInstanceByName(modelName string) (ModelRegistryInterface, error) {
// 	for modelType := range ModelRegistryMap {
// 		if modelType.Name() == modelName {
// 			return GetModelRegistryInstanceByType(modelType)
// 		}
// 	}
// 	return nil, fmt.Errorf("model %s not registered", modelName)
// }
