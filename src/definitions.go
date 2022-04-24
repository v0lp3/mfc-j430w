package main

type requests struct {
	leaseRequest string
	scanRequest  string
}

type modes struct {
	color     string
	grayscale string
}

type encode struct {
	jpeg string
	none string
}

type constants struct {
	ready       string
	mode        modes
	compression encode
	adfEnabled  int
	headerLen   int
	A4height    int
	mmInch      float32
	endPage     byte
	endScan     byte
}

var scanner constants = constants{
	ready: "+OK 200",
	mode: modes{
		color:     "CGRAY",
		grayscale: "GRAY64",
	},
	compression: encode{
		jpeg: "JPEG",
		none: "NONE",
	},
	adfEnabled: 0x2,
	headerLen:  0xc,
	A4height:   294,
	mmInch:     25.4,
	endPage:    0x82,
	endScan:    0x80,
}

var formats requests = requests{
	leaseRequest: "\x1bI\nR=%d,%d\nM=%s\n\x80",
	scanRequest:  "\x1bX\nR=%v,%v\nM=%s\nC=%s\nJ=MID\nB=50\nN=50\nA=0,0,%d,%d\n\x80",
}
