// Code generated by goctl. DO NOT EDIT.
package types

type APIRouteGroup struct {
	Jwt        bool        `json:"jwt,omitempty,optional"`
	Prefix     string      `json:"prefix,omitempty,optional"`
	Group      string      `json:"group,omitempty,optional"`
	Timeout    int         `json:"timeout,omitempty,optional,range=[0:]"`
	Middleware string      `json:"middleware,omitempty,optional"`
	MaxBytes   int64       `json:"maxBytes,omitempty,optional,range=[0:]"`
	Routes     []*APIRoute `json:"routes,omitempty"`
}

type APIRoute struct {
	Handler      string      `json:"handler,omitempty,optional"`
	Method       string      `json:"method,options=get|head|post|put|patch|delete|connect|options|trace"`
	Path         string      `json:"path"`
	ContentType  string      `json:"contentType,omitempty,optional,options=application/json|application/x-www-form-urlencoded"`
	RequestBody  []*FormItem `json:"requestBodyFields,omitempty,optional"`
	ResponseBody string      `json:"responseBody,omitempty,optional"`
}

type APIGenerateRequest struct {
	Name string           `json:"serviceName"`
	List []*APIRouteGroup `json:"routeGroups"`
}

type APIGenerateResponse struct {
	API string `json:"api"`
}

type FormItem struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Optional     bool   `json:"optional,omitempty"`
	DefaultValue string `json:"defaultValue,omitempty"`
	CheckEnum    bool   `json:"checkEnum,omitempty"`
	EnumValue    string `json:"enumValue,omitempty"`  // // effect if checkEunm is true
	LowerBound   int64  `json:"lowerBound,omitempty"` // effect if checkEunm is false
	UpperBound   int64  `json:"upperBound,omitempty"` // // effect if checkEunm is false
}

type ParseJsonRequest struct {
	JSON string `json:"json"`
}

type ParseJsonResponse struct {
	Form []*FormItem `json:"form"`
}
