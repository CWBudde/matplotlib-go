package color

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"io"

	"matplotlib-go/render"
)

const listedColormapSize = 256

// listedColormapRGB contains magma, inferno, plasma, and viridis RGB lookup
// tables from Matplotlib's _cm_listed.py, stored as 8-bit sRGB values.
const listedColormapRGB = "" +
	"eNpl0vVbmosCAGAURBEVcLqZIArY3Z0I6ow5GwO7ALuVsJWY0zmnmx24nfWOZ9tZ93bWXW4u1MVZ2MHHx3Pvc3+9z/P+CS8EAoVA" +
	"YBAIXA6iKCeHkJNDysurysujoFAMDLoFBtNUgG2DK2grKugqwfURijhlRUOkEkEFQVJVNlFTNkMhLdAq1hgVW3VVew1VR001l61q" +
	"bttQHloobx20ry46QA8diEUH4TAhBphQQ0yEESaSiIkmYWJNMPFmmERzTLIlmm6NTrNBZ9qhsx3QOY6oPGcUwxXFckcVeKCKvFDF" +
	"PqgSP1SpP6qMjCqnoCqo6MpgdFUIuioUUx2OqYlQr4lUr925pTZ6S12sRl2cJjtBk03byk7axk7exqFrcVK1OOnanEwdbpYON0eX" +
	"m6vHzdfjMvS5LCyvAMsrwvFKcLxSA145nleB51UZ8qoNebVGPDaBxyHweERePZHXSOI1kXgtxrw2E167CU9gyhWacneZcXebcTvN" +
	"uXvMud0W3B4LTq8lp8+Sc8CKM2DFHrRmD1uzR23YYzZ1E7Z1Ytu6g3a1f9jXHravOepQc9yh5oRj9SnH6imnqr+cqs44V551rjzn" +
	"UnHBteKia/llt7KrbmXX3EtveJTe9Ci57Vn8j1fxXa+i+96FD3wKH/kWPPZjPfVjPvNnvghgvCTnvw7Me0PJnabmvgvKeR+c/SEk" +
	"6+P2rE+hmbNhGXMR6fM70r5Epn2NSv0eTf83hv4jLuVnfMovWvKvpKTfyUkL9KTF1MSl9MTlzMTlbNpKDm01j7bKoK2xaGuFtPUi" +
	"2npJ4kZZ4kZF4mZV4mZN0mZdkoSTLOEmS+pTgEY60EwHWlOB9jSpIF0qSpd2ZEg7M6VdWWB3NtiTA/bmgvvzwP58cJABDrPAkQJw" +
	"rFA2USSbLJEdKpUdLpMdLZcdq5SdqJKdqpFN1cpO18nOcmTnuLLz9bKLDbLLTbKrzbLrrbKbbbJbfNkdoeyuCLzfAT7sBB93gU+6" +
	"wWc94It94Ks+8M0BcLoffDcIzgyDH0eln8elcxPS+Unp10PS74elP45Kfx2XLpyQLp6SLk9JV08D62eBzXOA5AIgvQT5//9yavLy" +
	"aKi8OgyqAYNuVYBpwWG6igr6SnAcAo5XVjRCKpJUlUzUlMxQCEs0wlpd2W6LsoMG0mkr0nUb0l0b6aWD9NFT8ddXIeNUKAYqwYbI" +
	"7QRkGAkZYYyMNEVGmSvHWCrHWSkn2Cgn2iGSHBApjgi6MyLVFZHmjsjwQGR6IbJ8EFm+iGx/5Ryycm6gci4VmReMzAtRyQ9VyQ9X" +
	"ZUSoMiLVGDvVmNEoZiyKGYdmJmCYNAwrSZ2Vos6ib2GlabDSNViZmqwsTVbOVlbeNlb+NhZTi8XSYhVqs4q1mSU6zDJdZrkus1KP" +
	"Wa3HrNFn1Okz2FgGF5fPw+U3GOQ3GeQ14/Na8Xlthrl8w1yhUa6IkNNByNlNzO4iZneTsvaSsvYZZ/aaZOw3yThgmj5gmj5kljZs" +
	"njpqnjpmQZ+wSBFbphy0Sv7DKumwdeJRG9oxW9oJ24STdvF/2sdN2ceedog54xjzt1P0OeeoC847L7pEXnLdccUt4qp7+HWPsBue" +
	"obe8tt/2CrnjHXzXJ+ieL/W+H+VhQOAjcsDjQP8nFL9nVN/nQT4vgr1fhni9DvV8E+b+NtxtOsL1XaTL+53OM1FOH2IcPsXaf46z" +
	"m02wnaPZzCdazSdbfkmx+Jpq/i3N7HuGyb+Zxj+yST9ziD/zjH4xDH8z8QsFuIVC7GKx/lKp3lKZznKF9nKV1kr11tVazVW2xhpX" +
	"fY2HWW9ArzepbbSobrSqbLQjNwWITZHSZoeSZLeipAsu6YZLeuBALxzYDwf64cCAIjCkCIwoAWPKwAQSmFQFDqkBh9HAUXXguAZw" +
	"YitwSguY0gFO6wFnsZJzeMkFI8klouSKyeY1s80bFpu3rDfu2G7cddi477z+0G39scfaU++1536rL8mrr6krb0NWpsOX30cuf4hZ" +
	"/hS/NJu0NE9f/Jqx+C1n8V/G4s/CxV+liwtVi0t1S8u8pZWm5bW2lXXh6sbuNUn3OtC7Ke2XgENS2ZiKYjNasWULvE0T3q4F5+vC" +
	"BfpwIU5BhFfYZaTQQVTYbazQaarQZa6wxxK2xxrWbQvbaw/rcYDtc4Ltc4H1usH63KF9ntD93tADPtADftD+AOgAGTpAkR8Mkh8M" +
	"lh/aLj8cJj8cLj+yQ24kUm40Sm40Rm40Vm4sXm4sATKeCBlPhoynQCZSIRNpkIkMiDgTIs6GiHMg4jyImAGZZEImCyCThZDJYshk" +
	"CWSyDDJZLjdZKTdZJTdZIz9ZKy9mQ8UcqJgHE9fDxI0KE03wiWbFiVbF8TalcT5iTKA8JlIZ26U62qE22oka6UIPd2OG96oP9WwZ" +
	"6tUY7NMc2L9toF+rf0D7wKDO/iHd/SN6faP6vWPYfeO4HjG+Z9Jw70Gj7kOEPYeJXUdInUeNdx8z6ThutuuE+a6TFqI/LYVTVoK/" +
	"rPmnbdrP2LWdtW/926HlnGPzeaemC86NF10aLrnWX3bnXfHgXvXkXvPiXPdm3/Cpu+lbe8uv5nZA9R1y1T+BlXcpFfeo5feDyh4E" +
	"lz0MKX0UWvI4rPhxeNGTiMKnOwqeRbKe72S+iGa+jGG8jM1/FZf3Oj73TULOW1r2dFLWdHLWu5TM9/SMmdT0mbS0DxmpHzPpn7Lo" +
	"n7NTPuckz+YlzeUnzjFo88yEL6yEr4XxX4vivhXHfi+J+V4a/W951I+KqB+VO39WRf6s2fGrNuJ3XfhvdtgCN3SBF7pYv32pMWSp" +
	"KXi5OWi5hbrSRllpp6zyA1eF5DVRwNou/7XdfuudvutdvhvdPht7vTf2eW32em72eUgOuEv63SSDbpIhV2DEBRh1BsacpBOOUrGj" +
	"9KCD9JC99LCd9IgteMwWPG4DnrQGT1mBU5bgX5bgGQvwrDl4zhw8bwZeNAUvm4JXTMBrJuB1Y/AmSXqLJL1Dkt4lSu8RpQ+IwCMC" +
	"8JgAPCVInhMkLwiSV4TN14TNt4SNd4SNGcL6B8L6J+LaLHFtjrj6hbT6jbTynbT8g7T8i7S0QFpaJC0ukxZWiQtrRr838AFyUQHy" +
	"0QHQWLJCHBmeQFaiBSISA5HJgSopgWqpFFQaBZNOUc+kaGRRNLMpW3MpWnlU7XyqDoOqx6LqF1CxhVSDIiq+mGpYQjUqpRDLKKRy" +
	"inEFxaSSYlZFMa+mWNQEWtYGWtcG2tQF2rLJdhyyA4fsyCU78QKceQGu9f5uDf7uDf4ejX6ejX5eTX4+Tb6+zb5+zT7+LT4BLd7k" +
	"Fm9Kqxe11SuozTO4zTOkzXN7m0dou0dYu3tEu/uOdrdIvttOvmsU3zWa7xLDd4kVOMcJnOMFTgkCJ5rAMVHgmCRwSBY6pAjt6UL7" +
	"VKFdmtAuXWibIbTNFNpmCW2yhTY5QutcoXWe0CpfZMUQWTFFliyRZYHIolBkUSSyKBaZl4jMS0VmZSKzcpFZhci0UmRaJTKp/p8a" +
	"kXGtyLhOZMwWkTgiEldE4omI9SJig5DQKCQ0CQnNQqMWoVGr0KhNaPhf7UJDvhAvEOKFArxIYLBLYNAhMNgtwHUKcF183B4+rpuP" +
	"28vH9rRj97Vje/+nrw27vw17oA3b34odaMUOtmKHWnDDLbiRZtxos8FYk8F4E36iCS9uNBQ3Gk02GB2sJxyqJ/7BIx3mGR/hmhzl" +
	"mB7jmB1nm5+oszhRZ3Wy1vpUjc2f1XZT1fZ/VTmernQ6U+Fyttz1bLn732We50q9zpf4XCj2u1gUcLGQfKmAcpkVdIUZcpURejU/" +
	"/Frujus5O29kR9/Mir2ZGX8rg3Y7LflOKv0OPe2flIy7Sdn3EnPv0fLvxzMfxBU+iCl+GF36KKriUWTV4x21T8LZT8J4T7c3PA1p" +
	"fhbU+pzCfx4ofBHQ8cK/86Vv90vvnldefa88+l+7Db52GXnjPP7GUfzW/tBbuyPTNsemrU6+s5h6Z376nenf740vvCddmiFcnTG8" +
	"MYO/9QH3zwfs/Q/6Dz/qPfmo++yjzstP2m8+aU1/0pr5rPXxs9bsZ60vs9rfZrV/zOr8mtNdnNNbnsOuzeM25vGSeUPwC+E/EkZZ" +
	"Cg=="

func init() {
	listed := mustDecodeListedColormaps(listedColormapRGB)
	for name, colors := range listed {
		colormaps[name] = Colormap{name: name, listed: colors}
	}
}

func mustDecodeListedColormaps(encoded string) map[string][]render.Color {
	compressed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		panic(err)
	}
	reader, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		panic(err)
	}
	defer reader.Close()
	raw, err := io.ReadAll(reader)
	if err != nil {
		panic(err)
	}

	names := []string{"magma", "inferno", "plasma", "viridis"}
	const bytesPerMap = listedColormapSize * 3
	if len(raw) != len(names)*bytesPerMap {
		panic("color: invalid listed colormap payload")
	}

	result := make(map[string][]render.Color, len(names))
	for i, name := range names {
		offset := i * bytesPerMap
		colors := make([]render.Color, listedColormapSize)
		for j := range colors {
			rgb := raw[offset+j*3:]
			colors[j] = render.Color{
				R: float64(rgb[0]) / 255,
				G: float64(rgb[1]) / 255,
				B: float64(rgb[2]) / 255,
				A: 1,
			}
		}
		result[name] = colors
	}
	return result
}
