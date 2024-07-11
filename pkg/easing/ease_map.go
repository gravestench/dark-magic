package easing

const (
	Power0 = "Power0"
	Power1 = "Power1"
	Power2 = "Power2"
	Power3 = "Power3"
	Power4 = "Power4"

	Linear = "Linear"

	Stepped = "Stepped"

	Quadratic      = "Quad"
	QuadraticIn    = "Quad.easeIn"
	QuadraticOut   = "Quad.easeOut"
	QuadraticInOut = "Quad.easeInOut"

	Cubic      = "Cubic"
	CubicIn    = "Cubic.easeIn"
	CubicOut   = "Cubic.easeOut"
	CubicInOut = "Cubic.easeInOut"

	Quartic      = "Quart"
	QuarticIn    = "Quart.easeIn"
	QuarticOut   = "Quart.easeOut"
	QuarticInOut = "Quart.easeInOut"

	Quintic      = "Quint"
	QuinticIn    = "Quint.easeIn"
	QuinticOut   = "Quint.easeOut"
	QuinticInOut = "Quint.easeInOut"

	Sine      = "Sine"
	SineIn    = "Sine.easeIn"
	SineOut   = "Sine.easeOut"
	SineInOut = "Sine.easeInOut"

	Exponential      = "Exponential"
	ExponentialIn    = "Exponential.easeIn"
	ExponentialOut   = "Exponential.easeOut"
	ExponentialInOut = "Exponential.easeInOut"

	Circular      = "Circular"
	CircularIn    = "Circular.easeIn"
	CircularOut   = "Circular.easeOut"
	CircularInOut = "Circular.easeInOut"

	Elastic      = "Elastic"
	ElasticIn    = "Elastic.easeIn"
	ElasticOut   = "Elastic.easeOut"
	ElasticInOut = "Elastic.easeInOut"

	Back      = "Back"
	BackIn    = "Back.easeIn"
	BackOut   = "Back.easeOut"
	BackInOut = "Back.easeInOut"

	Bounce      = "Bounce"
	BounceIn    = "Bounce.easeIn"
	BounceOut   = "Bounce.easeOut"
	BounceInOut = "Bounce.easeInOut"

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
