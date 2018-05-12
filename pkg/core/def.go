package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Bitspark/slang/pkg/utils"
	"strings"
	"regexp"
)

type InstanceDefList []*InstanceDef
type TypeDefMap map[string]*TypeDef
type Properties utils.MapStr
type Generics map[string]*TypeDef

type InstanceDef struct {
	Name       string     `json:"-" yaml:"-"`
	Operator   string     `json:"operator" yaml:"operator"`
	Properties Properties `json:"properties" yaml:"properties,omitempty"`
	Generics   Generics   `json:"generics" yaml:"generics,omitempty"`

	valid       bool
	OperatorDef OperatorDef `json:"-" yaml:"definition,omitempty"`
}

type OperatorDef struct {
	ServiceDefs  map[string]*ServiceDef  `json:"services,omitempty" yaml:"services,omitempty"`
	DelegateDefs map[string]*DelegateDef `json:"delegates,omitempty" yaml:"delegates,omitempty"`
	InstanceDefs InstanceDefList         `json:"operators,omitempty" yaml:"operators,omitempty"`
	PropertyDefs TypeDefMap              `json:"properties,omitempty" yaml:"properties,omitempty"`
	Connections  map[string][]string     `json:"connections,omitempty" yaml:"connections,omitempty"`
	Elementary   string                  `json:"-" yaml:"-"`

	valid bool
}

type DelegateDef struct {
	In  TypeDef `json:"in" yaml:"in"`
	Out TypeDef `json:"out" yaml:"out"`

	valid bool
}

type ServiceDef struct {
	In  TypeDef `json:"in" yaml:"in"`
	Out TypeDef `json:"out" yaml:"out"`

	valid bool
}

type TypeDef struct {
	// Type is one of "primitive", "number", "string", "boolean", "stream", "map", "generic"
	Type    string              `json:"type" yaml:"type"`
	Stream  *TypeDef            `json:"stream,omitempty" yaml:"stream,omitempty"`
	Map     map[string]*TypeDef `json:"map,omitempty" yaml:"map,omitempty"`
	Generic string              `json:"generic,omitempty" yaml:"generic,omitempty"`

	valid bool
}

// INSTANCE DEFINITION

func (d InstanceDef) Valid() bool {
	return d.valid
}

func (d *InstanceDef) Validate() error {
	if d.Name == "" {
		return fmt.Errorf(`instance name may not be empty`)
	}

	if strings.Contains(d.Name, " ") {
		return fmt.Errorf(`operator instance name may not contain spaces: "%s"`, d.Name)
	}

	if d.Operator == "" {
		return errors.New(`operator may not be empty`)
	}

	if strings.Contains(d.Operator, " ") {
		return fmt.Errorf(`operator may not contain spaces: "%s"`, d.Operator)
	}

	d.valid = true
	return nil
}

// OPERATOR DEFINITION

func (d OperatorDef) Valid() bool {
	return d.valid
}

func (d *OperatorDef) Validate() error {
	for _, srv := range d.ServiceDefs {
		if err := srv.Validate(); err != nil {
			return err
		}
	}

	for _, del := range d.DelegateDefs {
		if err := del.Validate(); err != nil {
			return err
		}
	}

	alreadyUsedInsNames := make(map[string]bool)
	for _, insDef := range d.InstanceDefs {
		if err := insDef.Validate(); err != nil {
			return err
		}

		if _, ok := alreadyUsedInsNames[insDef.Name]; ok {
			return fmt.Errorf(`colliding instance names within same parent operator: "%s"`, insDef.Name)
		}
		alreadyUsedInsNames[insDef.Name] = true
	}

	d.valid = true
	return nil
}

// SpecifyGenerics replaces generic types in the operator definition with the types given in the generics map.
// The values of the map are the according identifiers. It does not touch referenced values such as *TypeDef but
// replaces them with a reference on a copy.
func (d *OperatorDef) SpecifyGenericPorts(generics map[string]*TypeDef) error {
	srvs := make(map[string]*ServiceDef)
	for srvName := range d.ServiceDefs {
		srv := d.ServiceDefs[srvName].Copy()
		if err := srv.In.SpecifyGenerics(generics); err != nil {
			return err
		}
		if err := srv.Out.SpecifyGenerics(generics); err != nil {
			return err
		}
		srvs[srvName] = &srv
	}
	d.ServiceDefs = srvs

	dels := make(map[string]*DelegateDef)
	for delName := range d.DelegateDefs {
		del := d.DelegateDefs[delName].Copy()
		if err := del.In.SpecifyGenerics(generics); err != nil {
			return err
		}
		if err := del.Out.SpecifyGenerics(generics); err != nil {
			return err
		}
		dels[delName] = &del
	}
	d.DelegateDefs = dels
	for _, op := range d.InstanceDefs {
		for _, gp := range op.Generics {
			if err := gp.SpecifyGenerics(generics); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d OperatorDef) GenericsSpecified() error {
	for _, srv := range d.ServiceDefs {
		if err := srv.In.GenericsSpecified(); err != nil {
			return err
		}
		if err := srv.Out.GenericsSpecified(); err != nil {
			return err
		}
	}
	for _, del := range d.DelegateDefs {
		if err := del.In.GenericsSpecified(); err != nil {
			return err
		}
		if err := del.Out.GenericsSpecified(); err != nil {
			return err
		}
	}
	for _, op := range d.InstanceDefs {
		for _, gp := range op.Generics {
			if err := gp.GenericsSpecified(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d OperatorDef) Copy() OperatorDef {
	cpy := d
	cpy.InstanceDefs = nil
	cpy.Connections = nil

	cpy.ServiceDefs = make(map[string]*ServiceDef)
	for k, v := range d.ServiceDefs {
		c := v.Copy()
		cpy.ServiceDefs[k] = &c
	}

	cpy.DelegateDefs = make(map[string]*DelegateDef)
	for k, v := range d.DelegateDefs {
		c := v.Copy()
		cpy.DelegateDefs[k] = &c
	}

	cpy.PropertyDefs = make(map[string]*TypeDef)
	for k, v := range d.PropertyDefs {
		c := v.Copy()
		cpy.PropertyDefs[k] = &c
	}

	return cpy
}

func (def *OperatorDef) SpecifyOperator(gens Generics, props Properties) error {
	if !def.Valid() {
		err := def.Validate()
		if err != nil {
			return err
		}
	}

	for _, srv := range def.ServiceDefs {
		srv.In.SpecifyGenerics(gens)
		srv.Out.SpecifyGenerics(gens)
	}

	for _, dlg := range def.DelegateDefs {
		dlg.In.SpecifyGenerics(gens)
		dlg.Out.SpecifyGenerics(gens)
	}

	def.PropertyDefs.SpecifyGenerics(gens)

	props.Clean()

	for prop, propDef := range def.PropertyDefs {
		propVal, ok := props[prop]
		if !ok {
			return errors.New("Missing property " + prop)
		}
		if err := propDef.VerifyData(propVal); err != nil {
			return err
		}
	}

	newSrvs := make(map[string]*ServiceDef)
	for name, srv := range def.ServiceDefs {
		parsed, _ := ExpandExpression(name, props, def.PropertyDefs)
		for _, p := range parsed {
			srvCpy := &ServiceDef{}
			srvCpy.In = srv.In.Copy()
			srvCpy.In.ApplyProperties(props, def.PropertyDefs)
			srvCpy.Out = srv.Out.Copy()
			srvCpy.Out.ApplyProperties(props, def.PropertyDefs)
			newSrvs[p] = srvCpy
		}
	}
	def.ServiceDefs = newSrvs

	newDels := make(map[string]*DelegateDef)
	for name, dlg := range def.DelegateDefs {
		parsed, _ := ExpandExpression(name, props, def.PropertyDefs)
		for _, p := range parsed {
			dlgCpy := &DelegateDef{}
			dlgCpy.In = dlg.In.Copy()
			dlgCpy.In.ApplyProperties(props, def.PropertyDefs)
			dlgCpy.Out = dlg.Out.Copy()
			dlgCpy.Out.ApplyProperties(props, def.PropertyDefs)
			newDels[p] = dlgCpy
		}
	}
	def.DelegateDefs = newDels

	for _, childOpInsDef := range def.InstanceDefs {
		// Propagate property values to child operators
		for prop, propVal := range childOpInsDef.Properties {
			propKey, ok := propVal.(string)
			if !ok {
				continue
			}
			// Parameterized properties must start with a '$'
			if !strings.HasPrefix(propKey, "$") {
				continue
			}
			propKey = propKey[1:]
			if val, ok := props[propKey]; ok {
				childOpInsDef.Properties[prop] = val
			} else {
				return fmt.Errorf("unknown property \"%s\"", prop)
			}
		}

		for _, gen := range childOpInsDef.Generics {
			gen.SpecifyGenerics(gens)
		}

		err := childOpInsDef.OperatorDef.SpecifyOperator(childOpInsDef.Generics, childOpInsDef.Properties)
		if err != nil {
			return err
		}
	}

	def.PropertyDefs = nil

	return nil
}

// SERVICE DEFINITION

func (d *ServiceDef) Valid() bool {
	return d.valid
}

func (d *ServiceDef) Validate() error {
	if err := d.In.Validate(); err != nil {
		return err
	}

	if err := d.Out.Validate(); err != nil {
		return err
	}

	d.valid = true
	return nil
}

func (d ServiceDef) Copy() ServiceDef {
	cpy := ServiceDef{}

	cpy.In = d.In.Copy()
	cpy.Out = d.Out.Copy()

	return cpy
}

// DELEGATE DEFINITION

func (d *DelegateDef) Valid() bool {
	return d.valid
}

func (d *DelegateDef) Validate() error {
	if err := d.In.Validate(); err != nil {
		return err
	}

	if err := d.Out.Validate(); err != nil {
		return err
	}

	d.valid = true
	return nil
}

func (d DelegateDef) Copy() DelegateDef {
	cpy := DelegateDef{}

	cpy.In = d.In.Copy()
	cpy.Out = d.Out.Copy()

	return cpy
}

// TYPE DEFINITION

func (d TypeDef) Equals(p TypeDef) bool {
	if d.Type != p.Type {
		return false
	}

	if d.Type == "map" {
		if len(d.Map) != len(p.Map) {
			return false
		}

		for k, e := range d.Map {
			pe, ok := p.Map[k]
			if !ok {
				return false
			}
			if !e.Equals(*pe) {
				return false
			}
		}
	} else if d.Type == "stream" {
		if !d.Stream.Equals(*p.Stream) {
			return false
		}
	}

	return true
}

func (d *TypeDef) Valid() bool {
	return d.valid
}

func (d *TypeDef) Validate() error {
	if d.Type == "" {
		return errors.New("type must not be empty")
	}

	validTypes := []string{"generic", "primitive", "trigger", "number", "string", "binary", "boolean", "stream", "map"}
	found := false
	for _, t := range validTypes {
		if t == d.Type {
			found = true
			break
		}
	}
	if !found {
		return errors.New("unknown type")
	}

	if d.Type == "generic" {
		if d.Generic == "" {
			return errors.New("generic identifier missing")
		}
	} else if d.Type == "stream" {
		if d.Stream == nil {
			return errors.New("stream missing")
		}
		return d.Stream.Validate()
	} else if d.Type == "map" {
		if len(d.Map) == 0 {
			return errors.New("map missing or empty")
		}
		for _, e := range d.Map {
			if e == nil {
				return errors.New("map entry must not be null")
			}
			err := e.Validate()
			if err != nil {
				return err
			}
		}
	}

	d.valid = true
	return nil
}

func (d TypeDef) Copy() TypeDef {
	cpy := TypeDef{Type: d.Type, Generic: d.Generic}

	if d.Stream != nil {
		strCpy := d.Stream.Copy()
		cpy.Stream = &strCpy
	}
	if d.Map != nil {
		cpy.Map = make(map[string]*TypeDef)
		for k, e := range d.Map {
			subCpy := e.Copy()
			cpy.Map[k] = &subCpy
		}
	}

	return cpy
}

// SpecifyGenerics replaces generic types in the port definition with the types given in the generics map.
// The values of the map are the according identifiers. It does not touch referenced values such as *TypeDef but
// replaces them with a reference on a copy, which is very important to prevent unintended side effects.
func (d *TypeDef) SpecifyGenerics(generics map[string]*TypeDef) error {
	for identifier, pd := range generics {
		if d.Generic == identifier {
			*d = pd.Copy()
			return nil
		}

		if d.Type == "stream" {
			strCpy := d.Stream.Copy()
			d.Stream = &strCpy
			return strCpy.SpecifyGenerics(generics)
		} else if d.Type == "map" {
			mapCpy := make(map[string]*TypeDef)
			for k, e := range d.Map {
				eCpy := e.Copy()
				if err := eCpy.SpecifyGenerics(generics); err != nil {
					return err
				}
				mapCpy[k] = &eCpy
			}
			d.Map = mapCpy
		}
	}
	return nil
}

func (d TypeDef) GenericsSpecified() error {
	if d.Type == "generic" || d.Generic != "" {
		return errors.New("generic not replaced: " + d.Generic)
	}

	if d.Type == "stream" {
		return d.Stream.GenericsSpecified()
	} else if d.Type == "map" {
		for _, e := range d.Map {
			if err := e.GenericsSpecified(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d TypeDef) VerifyData(data interface{}) error {
	switch v := data.(type) {
	case nil:
		if d.Type == "stream" || d.Type == "primitive" || d.Type == "trigger" || d.Type == "string" || d.Type == "number" || d.Type == "boolean" {
			return nil
		}
	case string:
		if d.Type == "string" || d.Type == "primitive" || d.Type == "trigger" {
			return nil
		}
	case int:
		if d.Type == "number" || d.Type == "primitive" || d.Type == "trigger" {
			return nil
		}
	case float64:
		if d.Type == "number" || d.Type == "primitive" || d.Type == "trigger" {
			return nil
		}
	case bool:
		if d.Type == "boolean" || d.Type == "primitive" || d.Type == "trigger" {
			return nil
		}
	case map[string]interface{}:
		if d.Type == "trigger" {
			return nil
		}
		if d.Type == "map" {
			for k, sub := range d.Map {
				e, ok := v[k]
				if !ok {
					return errors.New("missing entry " + k)
				}
				if err := sub.VerifyData(e); err != nil {
					return err
				}
			}
			return nil
		}
	case []interface{}:
		if d.Type == "stream" {
			if d.Type == "trigger" {
				return nil
			}
			for _, v := range v {
				if err := d.Stream.VerifyData(v); err != nil {
					return err
				}
			}
			return nil
		}
	}

	return fmt.Errorf("exptected %s, got %v", d.Type, data)
}

// TYPE DEF MAP

func (t TypeDefMap) VerifyData(data map[string]interface{}) error {
	for k, v := range t {
		if _, ok := data[k]; !ok {
			return errors.New("missing entry " + k)
		}
		if err := v.VerifyData(data[k]); err != nil {
			return fmt.Errorf("%s: %s", k, err.Error())
		}
	}
	for k := range data {
		if _, ok := t[k]; !ok {
			return errors.New("unexpected " + k)
		}
	}
	return nil
}

func (t TypeDefMap) SpecifyGenerics(generics map[string]*TypeDef) error {
	for _, v := range t {
		if err := v.SpecifyGenerics(generics); err != nil {
			return err
		}
	}
	return nil
}

func (t TypeDefMap) GenericsSpecified() error {
	for k, v := range t {
		if err := v.GenericsSpecified(); err != nil {
			return fmt.Errorf("%s: %s", k, err.Error())
		}
	}

	return nil
}

func (d *TypeDef) ApplyProperties(props Properties, propDefs map[string]*TypeDef) error {
	if d.Type == "primitive" || d.Type == "string" || d.Type == "number" || d.Type == "boolean" || d.Type == "trigger" {
		return nil
	}
	var parsed []string
	if d.Type == "generic" {
		parsed, _ = ExpandExpression(d.Generic, props, propDefs)
		if len(parsed) != 1 {
			return errors.New("generic must be 1")
		}
		d.Generic = parsed[0]
	}
	if d.Type == "stream" {
		return d.Stream.ApplyProperties(props, propDefs)
	}
	if d.Type == "map" {
		newMap := make(map[string]*TypeDef)
		for k, v := range d.Map {
			parsed, _ = ExpandExpression(k, props, propDefs)
			for _, k2 := range parsed {
				vCpy := v.Copy()
				vCpy.ApplyProperties(props, propDefs)
				newMap[k2] = &vCpy
			}
		}
		d.Map = newMap
		return nil
	}
	return errors.New("unknown type " + d.Type)
}

// OPERATOR LIST MARSHALLING

func (ol *InstanceDefList) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	var im map[string]*InstanceDef
	if err := unmarshal(&im); err != nil {
		return err
	}

	instances := make([]*InstanceDef, len(im))
	i := 0
	for name, inst := range im {
		inst.Name = name
		instances[i] = inst
		i++
	}

	*ol = instances
	return nil
}

func (ol InstanceDefList) MarshalYAML() (interface{}, error) {
	im := make(map[string]*InstanceDef)
	for _, ins := range ol {
		im[ins.Name] = ins
	}
	return im, nil
}

func (ol *InstanceDefList) UnmarshalJSON(data []byte) error {
	var im map[string]*InstanceDef
	if err := json.Unmarshal(data, &im); err != nil {
		return err
	}

	instances := make([]*InstanceDef, len(im))
	i := 0
	for name, inst := range im {
		inst.Name = name
		instances[i] = inst
		i++
	}

	*ol = instances
	return nil
}

func (ol InstanceDefList) MarshalJSON() ([]byte, error) {
	im := make(map[string]*InstanceDef)
	for _, ins := range ol {
		im[ins.Name] = ins
	}
	return json.Marshal(im)
}

func (p Properties) Clean() {
	for k, v := range p {
		p[k] = utils.CleanValue(v)
	}
}

// PROPERTY PARSING

func expandExpressionPart(exprPart string, props Properties, propDefs map[string]*TypeDef) ([]string, error) {
	var vals []string
	prop, ok := props[exprPart]
	if !ok {
		return nil, errors.New("missing property " + exprPart)
	}
	propDef := propDefs[exprPart]
	if propDef.Type == "stream" {
		els := prop.([]interface{})
		for _, el := range els {
			vals = append(vals, fmt.Sprintf("%v", el))
		}
	} else {
		vals = []string{fmt.Sprintf("%v", prop)}
	}
	return vals, nil
}

func ExpandExpression(expr string, props Properties, propDefs map[string]*TypeDef) ([]string, error) {
	re := regexp.MustCompile("{(.*?)}")
	exprs := []string{expr}
	for _, expr := range exprs {
		parts := re.FindAllStringSubmatch(expr, -1)
		if len(parts) == 0 {
			break
		}
		for _, match := range parts {
			// This could be extended with more complex logic in the future
			vals, err := expandExpressionPart(match[1], props, propDefs)
			if err != nil {
				return nil, err
			}

			// Actual replacement
			var newExprs []string
			for _, val := range vals {
				for _, e := range exprs {
					newExprs = append(newExprs, strings.Replace(e, match[0], fmt.Sprintf("%s", val), 1))
				}
			}
			exprs = newExprs
		}
	}
	return exprs, nil
}