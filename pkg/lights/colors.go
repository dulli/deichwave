package lights

import "image/color"

type HSL struct {
	H, S, L float64
}

func hue_2_rgb(v1, v2, vH float64) float64 { //Function Hue_2_RGB
	if vH < 0 {
		vH += 1
	}
	if vH > 1 {
		vH -= 1
	}
	if (6 * vH) < 1 {
		return (v1 + (v2-v1)*6*vH)
	}
	if (2 * vH) < 1 {
		return v2
	}
	if (3 * vH) < 2 {
		return (v1 + (v2-v1)*((2.0/3.0)-vH)*6)
	}
	return v1
}

func (c *HSL) Get() color.Color {
	return c.GetWithAlpha(255)
}

func (c *HSL) GetWithAlpha(alpha uint8) color.Color {
	var r, g, b float64
	if c.S == 0 { //HSL from 0 to 1
		r = c.L * 255 //RGB results from 0 to 255
		g = c.L * 255
		b = c.L * 255
	} else {
		var v1, v2 float64
		if c.L < 0.5 {
			v2 = c.L * (1 + c.S)
		} else {
			v2 = (c.L + c.S) - (c.S * c.L)
		}

		v1 = 2*c.L - v2

		r = 255 * hue_2_rgb(v1, v2, c.H+(1.0/3.0))
		g = 255 * hue_2_rgb(v1, v2, c.H)
		b = 255 * hue_2_rgb(v1, v2, c.H-(1.0/3.0))
	}
	return color.NRGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: alpha,
	}
}

type Colormap []HSL

func ColormapRainbow(colors int) Colormap {
	hd := 1.0 / float64(colors-1)
	p := make([]HSL, colors)
	for i := range p {
		p[i] = HSL{H: float64(i) * hd, S: 1, L: 0.5}
	}
	return p
}
