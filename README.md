# Brother MFC-J430W WiFi scanner protocol

## Reasons

_Brother MFC-J430W has already scanner driver and you can download [here](https://support.brother.com/g/b/downloadtop.aspx?c=it&lang=it&prod=mfcj430w_all)_ **but that are prebuilt binary (x86/x64) and source code isn't public**. This is a problem if you want to use the scanner on ARM architecture, because if you don't have the source code of the driver you can't recompile it. Anyway this should work on every scanner that use `brscan4`, but I'm not sure.

## Scanning protocol

![protocol](./docs/protocol.png)

### Color Modes

- **GRAY64**: gray scale image
- **CGRAY**: color image
- **TEXT**: low resolution mode, **max output (304x434)**

### Compressions

- **JPEG**: JPEG compression
- **RLENGTH**: Run Length Encoding
- **NONE**: no compression

### Status codes

When we open a connection with the scanner on port `54921`, it responds with his status code:

- `+OK 200`: Ready to use
- `-NG 401`: Scanner is busy

### Lease

Now we can send a request that specify resolution and color mode, then the scanner sends to client an offer, based on request. The ADF is automatically forced when the scanner detects documents in the feeder. I called this part `lease` because it recalled me _DHCP lease_.


### Response

After we make a request, the scanner sends an offer to client. If we want continue we are ready to start scanning.

`dpi(x),dpi(y),adf,mm(x),px(x),mm(y),px(y)`

- `dpi`: dots per inch
- `adf`: automatic document feeder (1=enabled, 2=disabled)
- `mm`: millimeters (plane width and height)
- `pixels`: pixels (image width and height)

## Start scan

The request must contains the following payload:

- **R** = `dpi(x), dpi(y)`
- **M** = `color mode`
- **C** = `compression`
- **J** = MID (?)
- **B** = `brightness`
- **N** = `contrast`
- **A** = `x1,y1,x2,y2` (area)

**NOTE**: `x2` and `y2` are calculated from plane dimensions because _width_ received from response in [lease phase](#lease) is different from _width calculated_

As you can see the reversing isn't completed at all and i work on it rarely, but work in progress...

## Protocol implementation

There are still some work to do to implement all functionalities of the scanner, but the major part is done. In this moment the scanning with image compression isn't implemented.

### Compile

```bash
git clone https://github.com/v0lp3/mfc-j430w.git
go build -o mfc-j430w mfc-j430w/src/*.go
```

### Usage

```bash
./mfc-j430w --help
```

Output:

```bash
Usage of ./mfc-j430w:
  -a string
        IP address of the Brother scanner (default "192.168.0.157")
  -c string
        Color mode of the scan (CGRAY, GRAY64) (default "CGRAY")
  -m    Enable scan of all pages from feeder
  -n string
        Name of the output file (default "scan.jpg")
  -r int
        Resolution of the scan (default 300)
```

## Credits

[Andrea Maugeri](https://github.com/v0lp3)

Partially thanks to [this](https://github.com/davidar/mfc7400c/)
