package iopipe

type PluginInstantiator func(*wrapper) Plugin

type Plugin interface {
	RunHook(string)
	Meta() interface{}
	Name() string
}

const HOOK_PRE_SETUP = "pre:setup"
const HOOK_POST_SETUP = "post:setup"
const HOOK_PRE_INVOKE = "pre:invoke"
const HOOK_POST_INVOKE = "post:invoke"
const HOOK_PRE_REPORT = "pre:report"
const HOOK_POST_REPORT = "post:report"
