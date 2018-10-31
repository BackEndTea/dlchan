# DL chan

Downloads all images from a specified 4chan board.

## Usage

```bash
$ dlchan --board=v --out=./output
```

This will create a folde named `v` inside the `./output` folder, and save all images in that folder.
Images are named as they are on the website, which means that the unix timestap of when they were posted is the filename

## Why

This was mostly made to get a bit of a grip on Golang, as the following has to be done:

* Check command flags
* Access a JSON API
* Retrieve some values
* Check if files/directories exist
* Download images
* Save those to the filesystem
