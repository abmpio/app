package app

import "github.com/abmpio/abmp/pkg/factory"

// 用于post处理
type PostProcessor interface {
	AfterInitialization()
}

type postProcessor struct {
	factory    factory.InstantiateFactory
	subscribes []PostProcessor
}

func newPostProcessor(factory factory.InstantiateFactory) *postProcessor {
	return &postProcessor{
		factory: factory,
	}
}

var (
	postProcessors []interface{}
)

// RegisterPostProcessor register post processor
func RegisterPostProcessor(p ...interface{}) {
	postProcessors = append(postProcessors, p...)
}

// Init init the post processor
func (p *postProcessor) Init() {
	for _, processor := range postProcessors {
		ss, err := p.factory.InjectIntoFunc(nil, processor)
		if err == nil {
			p.subscribes = append(p.subscribes, ss.(PostProcessor))
		}
	}
}

// AfterInitialization post processor after initialization
func (p *postProcessor) AfterInitialization() {
	for _, processor := range p.subscribes {
		p.factory.InjectIntoFunc(nil, processor)
		processor.AfterInitialization()
	}
}

// PostProcessor的回调实现
type postProcessorFunc struct {
	initializationFunc func()
}

func (p *postProcessorFunc) AfterInitialization() {
	if p.initializationFunc == nil {
		return
	}
	p.initializationFunc()
}

// 使用函数创建一个PostProcessor对象
func NewPostProcessor(initializeFunc func()) PostProcessor {
	return &postProcessorFunc{
		initializationFunc: initializeFunc,
	}
}
