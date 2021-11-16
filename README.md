# Brother MFC-J430W protocol wrapper (wifi scanner)

## Reasons

_Brother MFC-J430W has already scanner driver and you can download [here](https://support.brother.com/g/b/downloadtop.aspx?c=it&lang=it&prod=mfcj430w_all)_ **but that are prebuilt binary (x86/x64) and source code isn't public**. My problem was that I wanted use scanner in my RPi4 but **those driver not works on ARM architecture**. In the end I solved this issue throught workaround using a vps... recently I resumed the project and I found a solution (partially thanks to [this](https://github.com/davidar/mfc7400c/)) and imho I think this is better than any workaround. I think that this should work on every scanner that use `brscan4`

## Scanner protocol

![protocol](./docs/protocol.png)

### Status codes

When we open a connection with the scanner on port 54921, it respond with his status code:

- `-401`: Scanner is busy
- `+OK 200`: Ready to use

### Lease

Now we can send a request that specify resolution and color mode, then scanner send to client a offer based on request.

Request:

```go
request := []byte(fmt.Sprintf("\x1bI\nR=%d,%d\nM=%s\n\x80", resolution, resolution, mode))
sendPacket(socket, request)
```

Response:

`300,300,2,209,2480,294,3472`

- `response[0]` `response[1]`: Resolution
- `response[3]` `response[5]`: Dimensions in mm
- `response[4]` `response[6]`: Dimensions in px
- `response[2]`: ?

**Color mode are**:

- **GRAY64**: gray scale image
- **CGRAY**: color image

Resolution are **100, 150, 300, 600, 1200, 2400**.
I called this part `leasing` because it recalled me _DHCP lease_

### Automatic document feeder

If specified it's possible to disable ADF and scan only one page.
Omit this if you want use ADF.

```go
if !adf {
  request = []byte("\x1bD\nADF\n\x80")
  sendPacket(socket, request)
  readPacket(socket)
}
```

## Start scan

Now we are ready to send start scan request:

```go
requestFormat := "\x1bX\nR=%v,%v\nM=%s\nC=%s\nJ=MID\nB=50\nN=50\nA=0,0,%v,%v\n\x80"
```

- **R** = `X_RES`, `Y_WIDTH`
- **M** = `CGRAY` or `GRAY64`
- **C** = `JPEG` or `RLENGTH` or `NONE`
- **J** = MID
- **B** = 50 (Brightness?)
- **N** = 50 (Contrast?)
- **A** = 0,0,`WIDTH`, `HEIGHT`

Documentation work in progress...

## Compile

```bash
git clone https://github.com/v0lp3/mfc-j430w.git
go build -o mfc-j430w mfc-j430w/src/*.go
```

## Usage

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

## To do

- [ ] Implement multi page scan for ADF
- [ ] Add flag to output compressed image