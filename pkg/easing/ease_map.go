package easing

const (
	Power0      = "Power0"
	Power1      = "Power1"
	Power2      = "Power2"
	Power3      = "Power3"
	Power4      = "Power4"
	Linear      = "Linear"
	Quadratic   = "Quad"
	Cubic       = "Cubic"
	Quartic     = "Quart"
	Quintic     = "Quint"
	Sine        = "Sine"
	Exponential = "Exponential"
	Circular    = "Circular"
	Elastic     = "Elastic"
	Back        = "Back"
	Bounce      = "Bounce"

	QuadraticIn   = "Quad.easeIn"
	CubicIn       = "Cubic.easeIn"
	QuarticIn     = "Quart.easeIn"
	QuinticIn     = "Quint.easeIn"
	SineIn        = "Sine.easeIn"
	ExponentialIn = "Exponential.easeIn"
	CircularIn    = "Circular.easeIn"
	ElasticIn     = "Elastic.easeIn"
	BackIn        = "Back.easeIn"
	BounceIn      = "Bounce.easeIn"

	QuadraticOut   = "Quad.easeOut"
	CubicOut       = "Cubic.easeOut"
	QuarticOut     = "Quart.easeOut"
	QuinticOut     = "Quint.easeOut"
	SineOut        = "Sine.easeOut"
	ExponentialOut = "Exponential.easeOut"
	CircularOut    = "Circular.easeOut"
	ElasticOut     = "Elastic.easeOut"
	BackOut        = "Back.easeOut"
	BounceOut      = "Bounce.easeOut"

	QuadraticInOut   = "Quad.easeInOut"
	CubicInOut       = "Cubic.easeInOut"
	QuarticInOut     = "Quart.easeInOut"
	QuinticInOut     = "Quint.easeInOut"
	SineInOut        = "Sine.easeInOut"
	ExponentialInOut = "Exponential.easeInOut"
	CircularInOut    = "Circular.easeInOut"
	ElasticInOut     = "Elastic.easeInOut"
	BackInOut        = "Back.easeInOut"
	BounceInOut      = "Bounce.easeInOut"

	Stepped = "Stepped"
	Default = Linear
)

var EaseMap = map[string]EaseFunctionProvider{
	Linear:  &LinearEaseProvider{},
	Bounce:  &BounceOutEaseProvider{},
	Stepped: &SteppedEaseProvider{},

	Power0: &LinearEaseProvider{},
	Power1: &QuadraticOutEaseProvider{},
	Power2: &CubicOutEaseProvider{},
	Power3: &QuarticOutEaseProvider{},
	Power4: &QuinticOutEaseProvider{},

	Quadratic:   &QuadraticOutEaseProvider{},
	Cubic:       &CubicOutEaseProvider{},
	Quartic:     &QuarticOutEaseProvider{},
	Quintic:     &QuinticOutEaseProvider{},
	Sine:        &SineOutEaseProvider{},
	Exponential: &ExponentialOutEaseProvider{},
	Circular:    &CircularOutEaseProvider{},
	Elastic:     &ElasticOutEaseProvider{},
	Back:        &BackOutEaseProvider{},

	QuadraticIn:   &QuadraticInEaseProvider{},
	CubicIn:       &CubicInEaseProvider{},
	QuarticIn:     &QuarticInEaseProvider{},
	QuinticIn:     &QuinticInEaseProvider{},
	SineIn:        &SineInEaseProvider{},
	ExponentialIn: &ExponentialInEaseProvider{},
	CircularIn:    &CircularInEaseProvider{},
	ElasticIn:     &ElasticInEaseProvider{},
	BackIn:        &BackInEaseProvider{},

	QuadraticOut:   &QuadraticOutEaseProvider{},
	CubicOut:       &CubicOutEaseProvider{},
	QuarticOut:     &QuarticOutEaseProvider{},
	QuinticOut:     &QuinticOutEaseProvider{},
	SineOut:        &SineOutEaseProvider{},
	ExponentialOut: &ExponentialOutEaseProvider{},
	CircularOut:    &CircularOutEaseProvider{},
	ElasticOut:     &ElasticOutEaseProvider{},
	BackOut:        &BackOutEaseProvider{},

	QuadraticInOut:   &QuadraticInOutEaseProvider{},
	CubicInOut:       &CubicInOutEaseProvider{},
	QuarticInOut:     &QuarticInOutEaseProvider{},
	QuinticInOut:     &QuinticInOutEaseProvider{},
	SineInOut:        &SineInOutEaseProvider{},
	ExponentialInOut: &ExponentialInOutEaseProvider{},
	CircularInOut:    &CircularInOutEaseProvider{},
	ElasticInOut:     &ElasticInOutEaseProvider{},
	BackInOut:        &BackInOutEaseProvider{},
}
