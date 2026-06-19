package cv

import (
	"bytes"
	"encoding/json"
	"sort"
)

// SkillsOrder returns the skill keys in their source order (as loaded), or
// sorted keys if no order was captured. Used by editors to preserve the order
// skills render in.
func (c *CV) SkillsOrder() []string {
	if len(c.skillsOrder) > 0 {
		return append([]string(nil), c.skillsOrder...)
	}
	keys := make([]string, 0, len(c.Skills))
	for k := range c.Skills {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// SetSkillsOrder records the desired skill key order for the next Marshal.
func (c *CV) SetSkillsOrder(order []string) {
	c.skillsOrder = append([]string(nil), order...)
}

// orderedSkills marshals a skills map with its keys in a fixed order, since
// encoding/json otherwise sorts map keys and would reshuffle the CV.
type orderedSkills struct {
	order []string
	m     map[string]string
}

func (o orderedSkills) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('{')
	first := true
	for _, k := range o.order {
		v, ok := o.m[k]
		if !ok {
			continue
		}
		if !first {
			b.WriteByte(',')
		}
		first = false
		kb, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		vb, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		b.Write(kb)
		b.WriteByte(':')
		b.Write(vb)
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

// Marshal returns indented JSON for the CV, emitting skills in the given key
// order (any map keys not listed are appended, sorted). Pass nil to use the
// CV's recorded order.
func (c *CV) Marshal(skillOrder []string) ([]byte, error) {
	if skillOrder == nil {
		skillOrder = c.SkillsOrder()
	}
	type alias CV // strip methods, avoid recursion
	aux := struct {
		*alias
		Skills orderedSkills `json:"skills"`
	}{
		alias:  (*alias)(c),
		Skills: orderedSkills{order: fullOrder(skillOrder, c.Skills), m: c.Skills},
	}
	return json.MarshalIndent(aux, "", "  ")
}

// fullOrder returns order followed by any map keys it omits, sorted.
func fullOrder(order []string, m map[string]string) []string {
	seen := make(map[string]bool, len(order))
	out := make([]string, 0, len(m))
	for _, k := range order {
		if _, ok := m[k]; ok && !seen[k] {
			seen[k] = true
			out = append(out, k)
		}
	}
	var rest []string
	for k := range m {
		if !seen[k] {
			rest = append(rest, k)
		}
	}
	sort.Strings(rest)
	return append(out, rest...)
}
