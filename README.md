# Wall-E
## Web Advanced Lossless and Lossy Enhancer

I started this project after watching the Wall-E movie and completing the Tour of Go.

The subtitle is obviously a joke as you will realize that it is neither advanced (after reading the source code) nor does it "enhances" the file. So come and try reducing the size of your "trashy" images.

The main objective of this project was practicing creating a web server, handling file uploads, calling external executables and better understanding the work of types and interfaces together.

This project does not implement encoding/decoding of images in Go. All it does is try to compress a file size using a library and some external commands (calling optipng, gifsicle and pngquant).

### Dependencies

- For png optimizing optipng (lossless) and pngquant (lossy) are needed.
- For gif optimizing gifsicle is needed

### Running this project

Clone the repository and run:

```
go run wall-e.go
```

A Webserver will start at localhost:8000

### Future Development

Other features that would increasing learning:

- [ ] Tests
- [ ] Security analysis
- [ ] Batch upload and optimization
- [ ] Distributed optimization using message queues
- [ ] Auto clean old uploads
- [ ] LocalStorage to remember past upload url and stats
- [ ] Use Reactjs practing (need more frontend feature to justify)

### Sources

[sources.txt](http://github.com/edusig/wall-e/blob/master/sources.txt) contains links of websites and repositories I've used during development, still missing some :)
