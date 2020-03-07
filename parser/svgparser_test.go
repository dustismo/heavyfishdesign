package parser

import (
	"strings"
	"testing"

	"github.com/dustismo/heavyfishdesign/dynmap"
	"github.com/dustismo/heavyfishdesign/path"
	"github.com/dustismo/heavyfishdesign/transforms"
	"github.com/dustismo/heavyfishdesign/util"
)

func TestSvg1(t *testing.T) {
	svg := `
	<?xml version="1.0" encoding="UTF-8" standalone="no"?><!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd"><svg width="498px" height="542px" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" xml:space="preserve" xmlns:serif="http://www.serif.com/" style="fill-rule:evenodd;clip-rule:evenodd;stroke-linecap:round;stroke-linejoin:round;stroke-miterlimit:1.5;"><path d="M496.507,496.235C490.879,435.814 389.208,419.738 388.823,425.661L388.699,350.407C362.901,353.626 276.978,336.906 275.936,268.792C276.978,200.678 362.901,183.959 388.699,187.177L388.823,111.924C389.208,117.847 490.879,101.771 496.507,41.35L468.606,0.5C467.563,68.614 381.641,85.334 355.843,82.115L355.719,157.369C355.334,151.446 253.663,167.522 248.035,227.942C242.408,167.522 140.737,151.446 140.352,157.369L140.227,82.115C114.43,85.334 28.507,68.614 27.465,0.5L0.5,41.35C6.128,101.771 107.798,117.847 108.184,111.924L108.308,187.177C134.106,183.959 220.028,200.678 221.07,268.792C220.028,336.906 134.106,353.626 108.308,350.407L108.184,425.661C107.798,419.738 6.128,435.814 0.5,496.235L27.465,541.146C28.507,473.032 114.43,456.312 140.227,459.531L140.352,384.277C140.737,390.2 242.408,374.124 248.035,313.704C253.663,374.124 355.334,390.2 355.719,384.277L355.843,459.531C381.641,456.312 467.563,473.032 468.606,541.146" style="fill:none;stroke:black;stroke-width:1px;"/><path d="M496.507,496.235C487.602,511.205 478.698,526.176 469.793,541.146" style="fill:none;stroke:black;stroke-width:1px;"/></svg>
	`
	p, err := SVGParser{}.ParseSVG(svg, util.NewLog())
	if err != nil {
		t.Errorf("Error %s", err.Error())
	}
	p, _ = transforms.ScaleTransform{
		Width:            5,
		SegmentOperators: path.NewSegmentOperators(),
	}.PathTransform(p)

	t.Errorf(path.SvgString(p, 3))
}

func TestParser3(t *testing.T) {
	svg := `
	<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg width="2016px" height="2016px" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" xml:space="preserve" xmlns:serif="http://www.serif.com/" style="fill-rule:evenodd;clip-rule:evenodd;stroke-linecap:round;stroke-linejoin:round;stroke-miterlimit:1.5;">
    <g id="Main-Design-1" serif:id="Main Design 1">
        <clipPath id="_clip1">
            <rect x="60" y="60" width="1062.39" height="974.416"/>
        </clipPath>
        <g clip-path="url(#_clip1)">
            <rect id="base_outer" x="60" y="60" width="1896" height="1896" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <path d="M1927.2,1663.2C1927.2,1636.71 1905.69,1615.2 1879.2,1615.2L1015.2,1615.2C988.708,1615.2 967.2,1636.71 967.2,1663.2L967.2,1879.2C967.2,1905.69 988.708,1927.2 1015.2,1927.2L1879.2,1927.2C1905.69,1927.2 1927.2,1905.69 1927.2,1879.2L1927.2,1663.2Z" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <path d="M1927.2,1351.2C1927.2,1324.71 1905.69,1303.2 1879.2,1303.2L1015.2,1303.2C988.708,1303.2 967.2,1324.71 967.2,1351.2L967.2,1538.4C967.2,1564.89 988.708,1586.4 1015.2,1586.4L1879.2,1586.4C1905.69,1586.4 1927.2,1564.89 1927.2,1538.4L1927.2,1351.2Z" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <path d="M1927.2,386.4C1927.2,359.908 1905.69,338.4 1879.2,338.4L929.947,338.4C903.455,338.4 881.947,359.908 881.947,386.4L881.947,573.6C881.947,600.092 903.455,621.6 929.947,621.6L1879.2,621.6C1905.69,621.6 1927.2,600.092 1927.2,573.6L1927.2,386.4Z" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <path d="M1927.2,386.4C1927.2,359.908 1905.69,338.4 1879.2,338.4L929.947,338.4C903.455,338.4 881.947,359.908 881.947,386.4L881.947,573.6C881.947,600.092 903.455,621.6 929.947,621.6L1879.2,621.6C1905.69,621.6 1927.2,600.092 1927.2,573.6L1927.2,386.4Z" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <path d="M1927.2,136.8C1927.2,110.308 1905.69,88.8 1879.2,88.8L399.043,88.8C372.551,88.8 351.043,110.308 351.043,136.8L351.043,261.6C351.043,288.092 372.551,309.6 399.043,309.6L1879.2,309.6C1905.69,309.6 1927.2,288.092 1927.2,261.6L1927.2,136.8Z" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <path d="M781.981,1351.2C781.981,1324.71 760.473,1303.2 733.981,1303.2L136.8,1303.2C110.308,1303.2 88.8,1324.71 88.8,1351.2L88.8,1587.1C88.8,1613.59 110.308,1635.1 136.8,1635.1L733.981,1635.1C760.473,1635.1 781.981,1613.59 781.981,1587.1L781.981,1351.2Z" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <path d="M781.981,1711.89C781.981,1685.4 760.473,1663.89 733.981,1663.89L136.8,1663.89C110.308,1663.89 88.8,1685.4 88.8,1711.89L88.8,1879.2C88.8,1905.69 110.308,1927.2 136.8,1927.2L733.981,1927.2C760.473,1927.2 781.981,1905.69 781.981,1879.2L781.981,1711.89Z" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <path d="M1927.2,698.4C1927.2,671.908 1905.69,650.4 1879.2,650.4L1227.96,650.4C1201.47,650.4 1179.96,671.908 1179.96,698.4L1179.96,945.6C1179.96,972.092 1201.47,993.6 1227.96,993.6L1879.2,993.6C1905.69,993.6 1927.2,972.092 1927.2,945.6L1927.2,698.4Z" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <path d="M1923.96,1070.4C1923.96,1043.91 1902.45,1022.4 1875.96,1022.4L1227.96,1022.4C1201.47,1022.4 1179.96,1043.91 1179.96,1070.4L1179.96,1226.4C1179.96,1252.89 1201.47,1274.4 1227.96,1274.4L1875.96,1274.4C1902.45,1274.4 1923.96,1252.89 1923.96,1226.4L1923.96,1070.4Z" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <path d="M1151.16,698.4C1151.16,671.908 1129.65,650.4 1103.16,650.4L915.547,650.4C889.055,650.4 867.547,671.908 867.547,698.4L867.547,1226.4C867.547,1252.89 889.055,1274.4 915.547,1274.4L1103.16,1274.4C1129.65,1274.4 1151.16,1252.89 1151.16,1226.4L1151.16,698.4Z" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <path d="M938.4,1351.2C938.4,1324.71 916.892,1303.2 890.4,1303.2L858.992,1303.2C832.5,1303.2 810.992,1324.71 810.992,1351.2L810.992,1879.2C810.992,1905.69 832.5,1927.2 858.992,1927.2L890.4,1927.2C916.892,1927.2 938.4,1905.69 938.4,1879.2L938.4,1351.2Z" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <path d="M320.905,136.8C320.905,110.308 299.397,88.8 272.905,88.8L136.8,88.8C110.308,88.8 88.8,110.308 88.8,136.8L88.8,573.6C88.8,600.092 110.308,621.6 136.8,621.6L272.905,621.6C299.397,621.6 320.905,600.092 320.905,573.6L320.905,136.8Z" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <path d="M853.147,386.4C853.147,359.908 831.638,338.4 805.147,338.4L399.043,338.4C372.551,338.4 351.043,359.908 351.043,386.4L351.043,573.6C351.043,600.092 372.551,621.6 399.043,621.6L805.147,621.6C831.638,621.6 853.147,600.092 853.147,573.6L853.147,386.4Z" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <path d="M838.747,698.4C838.747,671.908 817.238,650.4 790.747,650.4L543.895,650.4C517.403,650.4 495.895,671.908 495.895,698.4L495.895,1226.4C495.895,1252.89 517.403,1274.4 543.895,1274.4L790.747,1274.4C817.238,1274.4 838.747,1252.89 838.747,1226.4L838.747,698.4Z" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <path d="M467.095,698.4C467.095,671.908 445.587,650.4 419.095,650.4L136.8,650.4C110.308,650.4 88.8,671.908 88.8,698.4L88.8,1226.4C88.8,1252.89 110.308,1274.4 136.8,1274.4L419.095,1274.4C445.587,1274.4 467.095,1252.89 467.095,1226.4L467.095,698.4Z" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <circle id="screw_hole" cx="88.8" cy="88.8" r="12" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <circle id="screw_hole1" serif:id="screw_hole" cx="863.96" cy="633.6" r="12" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <circle id="screw_hole2" serif:id="screw_hole" cx="799.195" cy="1298.22" r="12" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <circle id="screw_hole3" serif:id="screw_hole" cx="1167.96" cy="1274.4" r="12" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <circle id="screw_hole4" serif:id="screw_hole" cx="1167.96" cy="643.435" r="12" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <circle id="screw_hole5" serif:id="screw_hole" cx="88.8" cy="1286.22" r="12" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <circle id="screw_hole6" serif:id="screw_hole" cx="88.8" cy="1930.38" r="12" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <circle id="screw_hole7" serif:id="screw_hole" cx="1927.2" cy="1930.38" r="12" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <circle id="screw_hole8" serif:id="screw_hole" cx="1931.52" cy="87.183" r="12" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <rect id="cut_line" x="60" y="60" width="1044.87" height="948" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <rect id="cut_line1" serif:id="cut_line" x="1104.87" y="60" width="851.134" height="948" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <rect id="cut_line2" serif:id="cut_line" x="1104.87" y="1008" width="851.134" height="948" style="fill:none;stroke:black;stroke-width:1.33px;"/>
            <rect id="cut_line3" serif:id="cut_line" x="60" y="1008" width="1044.87" height="948" style="fill:none;stroke:black;stroke-width:1.33px;"/>
        </g>
    </g>
</svg>

	`

	element, err := Parse(strings.NewReader(svg), true)

	p, err := ElementToPath(element, util.NewLog())
	if err != nil {
		t.Errorf("ERRP %s", err)
	}
	t.Errorf("RESULT: \n%s\n", path.SvgString(p, 3))
}

func TestParser1(t *testing.T) {
	var testCases = []struct {
		svg     string
		element Element
	}{
		{
			`
		<svg width="100" height="100">
			<circle cx="50" cy="50" r="40" fill="red" />
		</svg>
		`,
			Element{
				Name: "svg",
				Attributes: dynmap.Wrap(map[string]interface{}{
					"width":  "100",
					"height": "100",
				}),
				Children: []*Element{
					element("circle", map[string]interface{}{"cx": "50", "cy": "50", "r": "40", "fill": "red"}),
				},
			},
		},
		{
			`
		<svg height="400" width="450">
			<g stroke="black" stroke-width="3" fill="black">
				<path id="AB" d="M 100 350 L 150 -300" stroke="red" />
				<path id="BC" d="M 250 50 L 150 300" stroke="red" />
				<path d="M 175 200 L 150 0" stroke="green" />
			</g>
		</svg>
		`,
			Element{
				Name: "svg",
				Attributes: dynmap.Wrap(map[string]interface{}{
					"width":  "450",
					"height": "400",
				}),
				Children: []*Element{
					&Element{
						Name: "g",
						Attributes: dynmap.Wrap(map[string]interface{}{
							"stroke":       "black",
							"stroke-width": "3",
							"fill":         "black",
						}),
						Children: []*Element{
							element("path", map[string]interface{}{"id": "AB", "d": "M 100 350 L 150 -300", "stroke": "red"}),
							element("path", map[string]interface{}{"id": "BC", "d": "M 250 50 L 150 300", "stroke": "red"}),
							element("path", map[string]interface{}{"d": "M 175 200 L 150 0", "stroke": "green"}),
						},
					},
				},
			},
		},
		{
			"",
			Element{},
		},
	}

	for _, test := range testCases {
		actual, err := parse(test.svg, false)

		if !(test.element.Compare(actual) && err == nil) {
			t.Errorf("Parse: expected %v, actual %v\n", test.element, actual)
		}
	}
}

func TestValidDocument(t *testing.T) {
	svg := `
		<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" id="svg-root" width="100%" height="100%" viewBox="0 0 480 360">
			<title id="test-title">color-prop-01-b</title>
			<desc id="test-desc">Test that viewer has the basic capability to process the color property</desc>
			<rect id="test-frame" x="1" y="1" width="478" height="358" fill="none" stroke="#000000"/>
		</svg>
		`

	element, err := parse(svg, true)
	if !(element != nil && err == nil) {
		t.Errorf("Validation: expected %v, actual %v\n", nil, err)
	}
}

func element(name string, attrs map[string]interface{}) *Element {
	return &Element{
		Name:       name,
		Attributes: dynmap.Wrap(attrs),
		Children:   []*Element{},
	}
}

func parse(svg string, validate bool) (*Element, error) {
	element, err := Parse(strings.NewReader(svg), validate)
	return element, err
}
