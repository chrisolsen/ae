# image

## Overview

Wraps Google's cloud storage lib allowing easy saving and lazy resizing of images upon request using appengine. This library was created to allow for each personal code re-use when uploading images to GCS and was not intended to be super flexible.

No tests have been written yet, so use at your own risk.

## Usage

image provides methods to obtain `reader` and `writer` methods.

### Examples

**Note:** Error handling is omitted in the examples below

#### Read

```
reader, err := image.NewReader(ctx, "filename.jpg")
defer reader.Close()
var b []bytes
count, err := reader.Read(b)
```

#### Write

```
writer, err := image.NewWriter(ctx, "filename.jpg", "image/jpeg")
defer writer.Close()
count, err := writer.Write(dataFromRequest)
```

#### ResizedURL

```
http.HandleFunc("/images", func(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    name := r.URL.Query().Get("name")       // GCS file name
    w := r.URL.Query().Get("width")
    h := r.URL.Query().Get("height")

    width, err := strconv.Atoi(w)
    height, err := strconv.Atoi(h)

    url, err := image.ResizedURL(ctx, name, width, height)
    http.Redirect(w, r, url, http.StatusMovedPermanently)
})
```