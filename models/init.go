package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"

	"github.com/jinzhu/copier"

	"github.com/go-playground/validator/v10"
)

const (
	solidRegexString  = "^[a-z0-9_\\-]+$"
	clDateRegexString = "^[0-9]{2}[.-][0-9]{2}[.-][0-9]{4}$"
	semverRegexString = "^[0-9]+\\.[0-9]+(\\.[0-9]+)?$"
)

var validate *validator.Validate

// IValid is interface to control all models from user code
type IValid interface {
	Valid() error
}

// GetECSDefinitions is function to return Event Config Schema definitions defaults
func GetECSDefinitions(defs Definitions) map[string]*Type {
	eventTypes := []string{"atomic", "aggregation", "correlation"}
	if defs == nil {
		defs = make(Definitions, 0)
	}

	defs["actions"] = &Type{
		Type: "array",
		Items: &Type{
			Type: "object",
			Properties: map[string]*Type{
				"type": {
					Type: "string",
					Enum: []interface{}{"db"},
				},
				"name": {
					Type: "string",
					Enum: []interface{}{"log_to_db"},
				},
			},
			AdditionalProperties: []byte("false"),
			Required:             []string{"type", "name"},
		},
	}

	for _, eventType := range eventTypes {
		defs["types."+eventType] = &Type{
			Type:    "string",
			Default: eventType,
			Enum:    []interface{}{eventType},
		}
	}
	defs["events.atomic"] = &Type{
		Type: "object",
		Properties: map[string]*Type{
			"type": &Type{
				Ref: "#/definitions/types.atomic",
			},
			"actions": &Type{
				Ref: "#/definitions/actions",
			},
		},
		Required: []string{"type", "actions"},
	}
	defs["events.complex"] = &Type{
		Type: "object",
		Properties: map[string]*Type{
			"type": &Type{
				Type: "string",
			},
			"actions": &Type{
				Ref: "#/definitions/actions",
			},
			"seq": &Type{
				Type:     "array",
				MinItems: 1,
				Items: &Type{
					Type: "object",
					Properties: map[string]*Type{
						"name": &Type{
							Type: "string",
						},
						"min_count": &Type{
							Type:    "integer",
							Minimum: 1,
						},
					},
					Required: []string{"name", "min_count"},
				},
			},
			"group_by": &Type{
				Type:        "array",
				MinItems:    1,
				UniqueItems: true,
				Items: &Type{
					Type: "string",
				},
			},
			"max_count": &Type{
				Type:    "integer",
				Minimum: 0,
			},
			"max_time": &Type{
				Type:    "integer",
				Minimum: 0,
			},
		},
		Required: []string{
			"type",
			"actions",
			"seq",
			"group_by",
			"max_count",
			"max_time",
		},
	}
	defs["events.aggregation"] = &Type{
		AllOf: []*Type{
			&Type{
				Ref: "#/definitions/events.complex",
			},
			&Type{
				Type: "object",
				Properties: map[string]*Type{
					"type": &Type{
						Ref: "#/definitions/types.aggregation",
					},
					"seq": &Type{
						Type:     "array",
						MaxItems: 1,
					},
				},
				Required: []string{"type", "seq"},
			},
		},
	}
	defs["events.correlation"] = &Type{
		AllOf: []*Type{
			&Type{
				Ref: "#/definitions/events.complex",
			},
			&Type{
				Type: "object",
				Properties: map[string]*Type{
					"type": &Type{
						Ref: "#/definitions/types.correlation",
					},
					"seq": &Type{
						Type:     "array",
						MaxItems: 20,
					},
				},
				Required: []string{"type", "seq"},
			},
		},
	}

	return defs
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func templateValidatorString(regexpString string) validator.Func {
	regexpValue := regexp.MustCompile(regexpString)
	return func(fl validator.FieldLevel) bool {
		field := fl.Field()

		switch field.Kind() {
		case reflect.String:
			return regexpValue.MatchString(fl.Field().String())
		case reflect.Slice, reflect.Array:
			for i := 0; i < field.Len(); i++ {
				if !regexpValue.MatchString(field.Index(i).String()) {
					return false
				}
			}
			return true
		case reflect.Map:
			for _, k := range field.MapKeys() {
				if !regexpValue.MatchString(field.MapIndex(k).String()) {
					return false
				}
			}
			return true
		default:
			return false
		}
	}
}

func strongPasswordValidatorString() validator.Func {
	numberRegex := regexp.MustCompile("[0-9]")
	alphaLRegex := regexp.MustCompile("[a-z]")
	alphaURegex := regexp.MustCompile("[A-Z]")
	specRegex := regexp.MustCompile("[!@#$&*]")
	return func(fl validator.FieldLevel) bool {
		field := fl.Field()

		switch field.Kind() {
		case reflect.String:
			password := fl.Field().String()
			return len(password) > 15 || (len(password) >= 8 &&
				numberRegex.MatchString(password) &&
				alphaLRegex.MatchString(password) &&
				alphaURegex.MatchString(password) &&
				specRegex.MatchString(password))
		default:
			return false
		}
	}
}

func deepValidator() validator.Func {
	return func(fl validator.FieldLevel) bool {
		if iv, ok := fl.Field().Interface().(IValid); ok {
			if err := iv.Valid(); err != nil {
				return false
			}
		}

		return true
	}
}

func eventConfigItemStructValidator(sl validator.StructLevel) {
	eci, ok := sl.Current().Interface().(EventConfigItem)
	if !ok {
		return
	}

	if eci.Type != "atomic" {
		if eci.MaxCount == 0 && eci.MaxTime == 0 {
			// cannot be used infinite correlations and aggregations
			sl.ReportError(eci.MaxCount, "MaxCount", "max_count", "must_limit_event", "")
			sl.ReportError(eci.MaxTime, "MaxTime", "max_time", "must_limit_event", "")
		}
		if len(eci.Seq) == 0 {
			// cannot be used empty correlations and aggregations
			sl.ReportError(eci.Seq, "Seq", "seq", "must_seq_event", "")
		}
		if len(eci.GroupBy) == 0 {
			// cannot be used empty grouping on correlations and aggregations
			sl.ReportError(eci.GroupBy, "GroupBy", "group_by", "must_grouping_event", "")
		}
	}
}

func checkModuleTags(sl validator.StructLevel, mod ModuleA) {
	if len(mod.Info.Tags) != len(mod.Locale.Tags) {
		sl.ReportError(mod.Info.Tags, "Tags", "tags",
			"must_eq_len_locale_tags", "")
		sl.ReportError(mod.Locale.Tags, "Tags", "tags",
			"must_eq_len_locale_tags", "")
	} else {
		for _, tid := range mod.Info.Tags {
			if _, ok := mod.Locale.Tags[tid]; !ok {
				sl.ReportError(mod.Locale.Tags, "Tags", "tags",
					"must_val_in_locale_tags", "")
			}
		}
	}
}

func checkModuleEvents(sl validator.StructLevel, mod ModuleA) {
	if err := mod.EventDataSchema.Valid(); err != nil {
		sl.ReportError(mod.EventDataSchema, "EventDataSchema", "event_data_schema",
			"must_valid_event_data_schema", "")
	} else {
		for edid := range mod.EventDataSchema.Properties {
			if _, ok := mod.Locale.EventData[edid]; !ok {
				sl.ReportError(mod.Locale.EventData, "EventData", "event_data",
					"must_val_in_locale_event_data", "")
			}
		}
	}
	if len(mod.Info.Events) != len(mod.Locale.Events) {
		sl.ReportError(mod.Info.Events, "Events", "events",
			"must_eq_len_locale_event", "")
		sl.ReportError(mod.Locale.Events, "Events", "events",
			"must_eq_len_locale_event", "")
	} else if len(mod.Info.Events) != len(mod.DefaultEventConfig) {
		sl.ReportError(mod.Info.Events, "Events", "events",
			"must_eq_len_default_event_config", "")
		sl.ReportError(mod.DefaultEventConfig, "DefaultEventConfig",
			"default_event_config", "must_eq_len_default_event_config", "")
	} else if len(mod.Info.Events) != len(mod.CurrentEventConfig) {
		sl.ReportError(mod.Info.Events, "Events", "events",
			"must_eq_len_current_event_config", "")
		sl.ReportError(mod.CurrentEventConfig, "CurrentEventConfig",
			"current_event_config", "must_eq_len_current_event_config", "")
	} else {
		for _, eid := range mod.Info.Events {
			if _, ok := mod.Locale.Events[eid]; !ok {
				sl.ReportError(mod.Locale.Events, "Events", "events",
					"must_val_in_locale_event", "")
			}
			if ev, ok := mod.DefaultEventConfig[eid]; !ok {
				sl.ReportError(mod.DefaultEventConfig, "DefaultEventConfig",
					"default_event_config", "must_val_in_default_event_config", "")
			} else {
				if _, ok := mod.EventConfigSchema.Properties[eid]; ev.Type == "atomic" && !ok {
					sl.ReportError(mod.EventConfigSchema.Properties, "Properties",
						"properties", "must_val_in_event_config_schema_from_def", "")
				}
				if evcl, ok := mod.Locale.EventConfig[eid]; ev.Type == "atomic" && !ok {
					sl.ReportError(mod.Locale.EventConfig, "EventConfig",
						"event_config", "must_val_in_locale_event_config_from_def", "")
				} else {
					if len(ev.Config) != len(evcl) {
						sl.ReportError(mod.Locale.EventConfig, "EventConfig",
							"event_config", "must_eq_len_locale_event_config_from_def", "")
					} else {
						for evcid := range ev.Config {
							if _, ok := evcl[evcid]; !ok {
								sl.ReportError(mod.Locale.EventConfig, "EventConfig",
									"event_config", "must_opt_val_in_locale_event_config_from_def", "")
							}
						}
					}
				}
			}
			if ev, ok := mod.CurrentEventConfig[eid]; !ok {
				sl.ReportError(mod.CurrentEventConfig, "CurrentEventConfig",
					"current_event_config", "must_val_in_current_event_config", "")
			} else {
				if _, ok := mod.EventConfigSchema.Properties[eid]; ev.Type == "atomic" && !ok {
					sl.ReportError(mod.EventConfigSchema.Properties, "Properties",
						"properties", "must_val_in_event_config_schema_from_cur", "")
				}
				if evcl, ok := mod.Locale.EventConfig[eid]; ev.Type == "atomic" && !ok {
					sl.ReportError(mod.Locale.EventConfig, "EventConfig",
						"event_config", "must_val_in_locale_event_config_from_cur", "")
				} else {
					if len(ev.Config) != len(evcl) {
						sl.ReportError(mod.Locale.EventConfig, "EventConfig",
							"event_config", "must_eq_len_locale_event_config_from_cur", "")
					} else {
						for evcid := range ev.Config {
							if _, ok := evcl[evcid]; !ok {
								sl.ReportError(mod.Locale.EventConfig, "EventConfig",
									"event_config", "must_opt_val_in_locale_event_config_from_cur", "")
							}
						}
					}
				}
			}
		}

		eventConfigSchema := Schema{}
		copier.Copy(&eventConfigSchema, &mod.EventConfigSchema)
		eventConfigSchema.Definitions = GetECSDefinitions(eventConfigSchema.Definitions)

		if defEventConfig, err := json.Marshal(mod.DefaultEventConfig); err != nil {
			sl.ReportError(mod.DefaultEventConfig, "DefaultEventConfig",
				"default_event_config", "must_json_compile_default_event_config", "")
		} else {
			if res, err := eventConfigSchema.ValidateBytes(defEventConfig); err != nil {
				sl.ReportError(mod.DefaultEventConfig, "DefaultEventConfig",
					"default_event_config", "must_valid_default_event_config_by_check", "")
			} else if !res.Valid() {
				sl.ReportError(mod.DefaultEventConfig, "DefaultEventConfig",
					"default_event_config", "must_valid_default_event_config_by_schema", "")
				fmt.Println("Errors: ", res.Errors())
			}
		}
		if curEventConfig, err := json.Marshal(mod.CurrentEventConfig); err != nil {
			sl.ReportError(mod.CurrentEventConfig, "CurrentEventConfig",
				"current_event_config", "must_json_compile_current_event_config", "")
		} else {
			if res, err := eventConfigSchema.ValidateBytes(curEventConfig); err != nil {
				sl.ReportError(mod.CurrentEventConfig, "CurrentEventConfig",
					"current_event_config", "must_valid_current_event_config_by_check", "")
			} else if !res.Valid() {
				sl.ReportError(mod.CurrentEventConfig, "CurrentEventConfig",
					"current_event_config", "must_valid_current_event_config_by_schema", "")
			}
		}
	}
}

func checkModuleConfig(sl validator.StructLevel, mod ModuleA) {
	if err := mod.ConfigSchema.Valid(); err != nil {
		sl.ReportError(mod.ConfigSchema, "ConfigSchema", "config_schema",
			"must_valid_config_schema", "")
	}
	if len(mod.CurrentConfig) != len(mod.Locale.Config) {
		sl.ReportError(mod.CurrentConfig, "CurrentConfig", "current_config",
			"must_eq_len_locale_event_config_from_cur", "")
		sl.ReportError(mod.Locale.Config, "Config", "config",
			"must_eq_len_locale_event_config_from_cur", "")
	} else {
		for cid := range mod.CurrentConfig {
			if _, ok := mod.Locale.Config[cid]; !ok {
				sl.ReportError(mod.Locale.Config, "Config", "config",
					"must_val_in_locale_config_from_cur", "")
			}
		}
	}
	if len(mod.DefaultConfig) != len(mod.Locale.Config) {
		sl.ReportError(mod.DefaultConfig, "DefaultConfig", "default_config",
			"must_eq_len_locale_event_config_from_def", "")
		sl.ReportError(mod.Locale.Config, "Config", "config",
			"must_eq_len_locale_event_config_from_def", "")
	} else {
		for cid := range mod.DefaultConfig {
			if _, ok := mod.Locale.Config[cid]; !ok {
				sl.ReportError(mod.Locale.Config, "Config", "config",
					"must_val_in_locale_config_from_def", "")
			}
		}
	}
	if curModuleConfig, err := json.Marshal(mod.CurrentConfig); err != nil {
		sl.ReportError(mod.CurrentConfig, "CurrentConfig",
			"current_config", "must_json_compile_current_config", "")
	} else {
		if res, err := mod.ConfigSchema.ValidateBytes(curModuleConfig); err != nil {
			sl.ReportError(mod.CurrentConfig, "CurrentConfig",
				"current_config", "must_valid_current_config_by_check", "")
		} else if !res.Valid() {
			sl.ReportError(mod.CurrentConfig, "CurrentConfig",
				"current_config", "must_valid_current_config_by_schema", "")
		}
	}
	if defModuleConfig, err := json.Marshal(mod.DefaultConfig); err != nil {
		sl.ReportError(mod.DefaultConfig, "DefaultConfig",
			"default_config", "must_json_compile_default_config", "")
	} else {
		if res, err := mod.ConfigSchema.ValidateBytes(defModuleConfig); err != nil {
			sl.ReportError(mod.DefaultConfig, "DefaultConfig",
				"default_config", "must_valid_default_config_by_check", "")
		} else if !res.Valid() {
			sl.ReportError(mod.DefaultConfig, "DefaultConfig",
				"default_config", "must_valid_default_config_by_schema", "")
		}
	}
}

func checkModuleVersion(sl validator.StructLevel, mod ModuleA) {
	if _, ok := mod.Changelog[mod.Info.Version]; !ok {
		sl.ReportError(mod.Info.Version, "Version", "version",
			"must_mod_version_in_changelog", "")
		sl.ReportError(mod.Changelog, "Changelog", "changelog",
			"must_mod_version_in_changelog", "")
	}
}

func systemModuleStructValidator(sl validator.StructLevel) {
	modS, ok := sl.Current().Interface().(ModuleS)
	if !ok {
		return
	}

	modA := modS.ToModuleA()
	checkModuleTags(sl, modA)
	checkModuleEvents(sl, modA)
	checkModuleConfig(sl, modA)
	checkModuleVersion(sl, modA)
}

func aagentModuleStructValidator(sl validator.StructLevel) {
	modA, ok := sl.Current().Interface().(ModuleA)
	if !ok {
		return
	}

	checkModuleTags(sl, modA)
	checkModuleEvents(sl, modA)
	checkModuleConfig(sl, modA)
	checkModuleVersion(sl, modA)
}

func init() {
	validate = validator.New()
	validate.RegisterValidation("solid", templateValidatorString(solidRegexString))
	validate.RegisterValidation("cldate", templateValidatorString(clDateRegexString))
	validate.RegisterValidation("semver", templateValidatorString(semverRegexString))
	validate.RegisterValidation("stpass", strongPasswordValidatorString())
	validate.RegisterValidation("valid", deepValidator())
	validate.RegisterStructValidation(eventConfigItemStructValidator, EventConfigItem{})
	validate.RegisterStructValidation(systemModuleStructValidator, ModuleS{})
	validate.RegisterStructValidation(aagentModuleStructValidator, ModuleA{})

	// Check validation interface for all models
	_, _ = reflect.ValueOf(Schema{}).Interface().(IValid)

	_, _ = reflect.ValueOf(User{}).Interface().(IValid)
	_, _ = reflect.ValueOf(Password{}).Interface().(IValid)
	_, _ = reflect.ValueOf(UserGroup{}).Interface().(IValid)
	_, _ = reflect.ValueOf(UserTenant{}).Interface().(IValid)
	_, _ = reflect.ValueOf(UserGroupTenant{}).Interface().(IValid)

	_, _ = reflect.ValueOf(Group{}).Interface().(IValid)

	_, _ = reflect.ValueOf(Tenant{}).Interface().(IValid)

	_, _ = reflect.ValueOf(ServiceInfoDB{}).Interface().(IValid)
	_, _ = reflect.ValueOf(ServiceInfoS3{}).Interface().(IValid)
	_, _ = reflect.ValueOf(ServiceInfoServer{}).Interface().(IValid)
	_, _ = reflect.ValueOf(ServiceInfo{}).Interface().(IValid)
	_, _ = reflect.ValueOf(Service{}).Interface().(IValid)
	_, _ = reflect.ValueOf(ServiceTenant{}).Interface().(IValid)

	_, _ = reflect.ValueOf(ModuleConfig{}).Interface().(IValid)
	_, _ = reflect.ValueOf(EventConfigAction{}).Interface().(IValid)
	_, _ = reflect.ValueOf(EventConfigSeq{}).Interface().(IValid)
	_, _ = reflect.ValueOf(EventConfigItem{}).Interface().(IValid)
	_, _ = reflect.ValueOf(EventConfig{}).Interface().(IValid)
	_, _ = reflect.ValueOf(ChangelogDesc{}).Interface().(IValid)
	_, _ = reflect.ValueOf(Changelog{}).Interface().(IValid)
	_, _ = reflect.ValueOf(LocaleDesc{}).Interface().(IValid)
	_, _ = reflect.ValueOf(Locale{}).Interface().(IValid)
	_, _ = reflect.ValueOf(ModuleInfo{}).Interface().(IValid)
	_, _ = reflect.ValueOf(ModuleS{}).Interface().(IValid)
	_, _ = reflect.ValueOf(ModuleSTenant{}).Interface().(IValid)
	_, _ = reflect.ValueOf(ModuleA{}).Interface().(IValid)
	_, _ = reflect.ValueOf(ModuleAAgent{}).Interface().(IValid)

	_, _ = reflect.ValueOf(AgentOS{}).Interface().(IValid)
	_, _ = reflect.ValueOf(AgentUser{}).Interface().(IValid)
	_, _ = reflect.ValueOf(AgentInfo{}).Interface().(IValid)
	_, _ = reflect.ValueOf(Agent{}).Interface().(IValid)

	_, _ = reflect.ValueOf(EventInfo{}).Interface().(IValid)
	_, _ = reflect.ValueOf(Event{}).Interface().(IValid)
	_, _ = reflect.ValueOf(EventModuleA{}).Interface().(IValid)
	_, _ = reflect.ValueOf(EventAgent{}).Interface().(IValid)
	_, _ = reflect.ValueOf(EventModuleAAgent{}).Interface().(IValid)
}
