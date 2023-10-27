# Bulk Uploading to Immich with `immich-go`

You may have an extended collection of personal photos and videos accumulated during years and stored in several folders, a NAS, on your own Google Photos account, or on partner's account. And you want to move them to an Immich server you own. 

The `immich-go` tool will help you to import massive Google Photos archives and your NAS folder:

- import from folder(s).
- import from zipped archives without prior extraction.
- discard duplicate images, based on the file name, and the date of capture.
- import only missing files or better files (an delete the inferior copy from the server).
- import from Google Photos takeout archives:
    - use metadata to  bypass file name discrepancies in the archive
    - use metadata to get album real names
    - use date of capture found in the json files
- import photos taken within a date range.
- create albums based on Google Photos albums or folder names.
- remove duplicated assets, based on the file name, date of capture, and file size
- no installation, no dependencies.

⚠️ This an early version, not yet extensively tested<br>
⚠️ Keep a backup copy of your files for safety<br>

For insights into the reasoning behind this alternative to `immich-cli`, please refer to the [Motivation](#motivation) section below.


# The Google photos takeout case

This project aims to make the process of importing Google photos takeouts as easy and accurate as possible. But keep in mind that 
Google takeout structure is complex and not documented. Some information may miss or even is wrong. 

## Here is the structure of the takeout archives I have encountered while doing this project

### Folders
  - The Year folder contains all image taken that year
  - Albums are in separate folders named as the album
  - album's actual name in stored into a metadata json file
  - not always, the title can be empty.
  - the json file name is in the user's language
  - The trash folder is names in the user's language
  - Hopefully, the JSON has a Trashed field.
  - The "Failed Videos" contains unreadable videos

### Images have a companion JSON file
  - the JSON contains some information on the image
  - The title has the original image name (that can be totally different of the image name in the archive)
  - the date of capture (epoch)
  - the GPS coordinates
  - Trashed flag
  - Partner flag

### The JSON file and the image name matches with some weird rules
  - The name length of the image can be shorter by 1 char compared to the name of the JSON.

### 2+ different images having the same name taken the same year are placed into the same folder with a number
  - IMG_3479.JPG
  - IMG_3479(1).JPG
  - IMG_3479(2).JPG

#### In that case, the JSONs are named:
  - IMG_3479.JPG.json
  - IMG_3479.JPG(1).json
  - IMG_3479.JPG(2).json

### Edited images may not have corresponding JSON.
  - PXL_20220405_090123740.PORTRAIT.jpg
  - PXL_20220405_090123740.PORTRAIT-modifié.jpg
but one JSON
  - PXL_20220405_090123740.PORTRAIT.jpg.json
Note that edited name is localized.


## What if you have problems with a takeout archive?
Please open an issue with details. You cna share your files using Discord DM @`simulot`.
I'll check if I can improve the program.
Sometime a manual import is the best option.


# Installation from release:

Installing `immich-go` is a straightforward process. Visit the [latest release page](https://github.com/simulot/immich-go/releases/latest) and select the binary file compatible with your system:

- Darwin arm-64, x86-64
- Linux arm-64, x86-64, i386
- Windows arm-64, x86-64, i386

The installation process involves copying and decompressing the corresponding binary executable onto your system. Open a shell and run the command `immich-go`.


⚠️ Please note that the linux x86-64 version is the only one tested.


# Installation from sources

For a source-based installation, ensure you have the necessary Go language development tools (https://go.dev/doc/install) in place.
Download the source files or clone the repository. 

Notably, there are no external dependencies to complicate the process.


# Executing `immich-go`
The `immich-go` program uses the Immich API. Hence it need the server address and a valid API key.


```sh
immich-go -server URL -key KEY -general_options COMMAND -command_options... {files}
```

`-server URL` URL of the Immich service, example http://<your-ip>:2283 or https://your-domain<br>
`-key KEY` A key generated by the user. Uploaded photos will belong to the key's owner.<br>
`-no-colors-log` Remove color codes from logs.<br>
`-log-level` Adjust the log verbosity as follow: (Default OK)
- `ERROR`: Display only errors
- `WARNING`: Same as previous one plus non blocking error
- `OK`: Same as previous plus actions
- `INFO`: Same as previous one plus progressions


## Command `upload`

Use this command for uploading photos and videos from a local directory, a zipped folder or all zip files that google photo takeout procedure has generated.
### Switches and options:
`-album "ALBUM NAME"` Import assets into the Immich album `ALBUM NAME`.<br>
`-device-uuid VALUE` Force the device identification (default $HOSTNAME).<br>
`-dry-run` Preview all actions as they would be done.<br> 
`-delete` Delete local assets after successful upload. <br>
`-create-album-folder <bool>` Generate immich albums after folder names.<br>
`-force-sidecar <bool>` Force sending a .xmp sidecar file beside images. With Google photos date and GPS coordinates are taken from metadata.json files. (default: FALSE).<br>


### Date selection:
Fine-tune import based on specific dates:<br>
`-date YYYY-MM-DD` import photos taken on a particular day.<br>
`-date YYYY-MM` select photos taken during a particular month.<br>
`-date YYYY` select photos taken during a particular year.<br>
`-date YYYY-MM-DD,YYYY-MM-DD` select photos taken within this date range.<br>

### Google photos options:

Specialized options for Google Photos management:
`-google-photos` import from a Google Photos structured archive, recreating corresponding albums.<br>
`-from-album "GP Album"` import assets for the given album, and mirrors it in Immich.<br>
`-create-albums <bool>`  Controls recreation of Google Photos albums in Immich (default: TRUE).<br>
`-keep-partner <bool>` Specifies inclusion or exclusion of partner-taken photos (default: TRUE).<br>
`-partner-album "partner's album"` import assets from partner into given album.<br>
`-keep-trashed <bool>` Determines whether to import trashed images (default: FALSE).<br>


### Example Usage: uploading a Google photos takeout archive

To illustrate, here's a command importing photos from a Google Photos takeout archive captured between June 1st and June 30th, 2019, while auto-generating albums:

```sh
./immich-go -server=http://mynas:2283 -key=zzV6k65KGLNB9mpGeri9n8Jk1VaNGHSCdoH1dY8jQ upload
-create-albums -google-photos -date=2019-06 ~/Download/takeout-*.zip             
```

## Command `duplicate`

Use this command for analyzing the content of your `immich` server to find any files that share the same file name, the  date of capture, but having different size. 
Before deleting the inferior copies, the system get all albums they belong to, and add the superior copy to them.

### Switches and options:
`-yes` Assume Yes to all questions (default: FALSE).<br> 
`-date` Check only assets have a date of capture in the given range. (default: 1850-01-04,2030-01-01)


### Example Usage: clean the `immich` server after having merged a google photo archive and original files

This command examine the immich server content, remove less quality images, and preserve albums.

NOTE: You should disable the dry run mode explicitly.

```sh
./immich-go -server=http://mynas:2283 -key=zzV6k65KGLNB9mpGeri9n8Jk1VaNGHSCdoH1dY8jQ duplicate -dry-run=false
```


# Merging strategy

The local file is analyzed to get following data:
- file size in bytes
- date of capture took from the takeout metadata, the exif data, or the file name with possible. The key is made of the file name + the size in the same way used by the immich server.

Digital cameras often generate file names with a sequence of 4 digits, leading to generate duplicated names. If the names matches, the capture date must be compared.

Tests are done in this order
1. the key is found in immich --> the name and the size match. We have the file, don't upload it.
1. the file name is found in immich and...
    1. dates match and immich file is smaller than the file --> Upload it, and discard the inferior file
    1. dates match and immich file is bigger than the file --> We have already a better version. Don't upload the file. 
1. Immich don't have it. --> Upload the file.
1. Update albums


# Acknowledgments

Kudos to the Immich team for they stunning project

This program use following 3rd party libraries:
- github.com/gabriel-vasile/mimetype to get exact file type
- github.com/rwcarlsen/goexif to get date of capture from JPEG files
- github.com/ttacon/chalk for having logs nicely colored 



# Motivation

The Immich project fulfills all my requirements for managing my photos:

- Self-hosted
- Open source
- Abundant functionalities
- User experience closely resembling Google Photos
- Machine learning capabilities
- Well-documented API
- Includes an import tool
- Continuously enhanced
- ...

Now, I need to migrate my photos to the new system in bulk. Most of my photos are stored in a NAS directory, while photos taken with my smartphone are in the Google Photos application often more compressed.

To completely transition away from the Google Photos service, I must set up an Immich server, import my NAS-stored photos, and merge them with my Google Photos collection. 
However, there are instances where the same pictures exist in both systems, sometimes with varying quality. Of course, I want to keep only the best copy of the photo.

The  `immich-cli` installation isn't trivial on a client machine, and doesn't handle Google Photos Takeout archive oddities.

The immich-cli tool does a great for importing a tone of files at full speed. However, I want more. So I write this utility for my onw purpose. Maybe, it could help some one else.

## Limitations of the `immich-CLI`:

While the provided tool is very fast and do the job, certain limitations persist:

### Advanced Expertise Required

The CLI tool is available within the Immich server container, eliminating the need to install the `Node.js` tool belt on your PC. Editing the `docker-compose.yml` file is necessary to access the host's files and retrieve your photos. Uploading photos from a different PC than the Immich server requires advanced skills.

### Limitations with Google Takeout Data

The Google Photos Takeout service delivers your collection as massive zip files containing your photos and their JSON files.

After unzipping the archive, you can use the CLI tool to upload its contents. However, certain limitations exist:
- Photos are organized in folders by year and albums.
- Photos are duplicated across year folders and albums.
- Some folders aren't albums
- Photos might be compressed in the takeout archive, affecting the CLI's duplicate detection when comparing them to previously imported photos with finest details.
- File and album names are mangled and correct names are found in JSON files

## Why the GO language?

The main reason is that my higher proficiency in GO compared to Typescript language.
Additionally, deploying a Node.js program on user machines presents challenges.

# Feature list
- [X] binary releases
- [ ] import vs upload flag
- [X] check in the photo doesn't exist on the server before uploading
    - [X] but keep files with the same name: ex IMG_0201.jpg if they aren't duplicates
    - [ ] some files may have different names (ex IMG_00195.jpg and IMAGE_00195 (1).jpg) and are true duplicates
- [X] replace the server photo, if the file to upload is better.
    - [X] Update any album with the new version of the asset
- [X] delete local file after successful upload (not for import!)
- [X] upload XMP sidecar files 
- [ ] select or exclude assets to upload by
    - [ ] type photo / video
    - [ ] name pattern
    - [ ] glob expression like ~/photos/\*/sorted/*.*
    - [ ] size
    - [X] date of capture within a date range
- [ ] multithreaded 
- [X] import from local folder
    - [X] create albums based on folder
    - [X] create an album with a given name
- [X] import from zip archives without unzipping them
- [X] Import Google takeout archive
    - [X] handle archives without unzipping them
    - [X] manage multi-zip archives
    - [X] Replicate google albums in immich
    - [X] manage duplicates assets inside the archive
    - [X] Don't upload google file if the server image is better
    - [X] don't import trashed files
    - [X] don't import failed videos
    - [ ] handle Archives 
    - [X] option to include photos taken by a partner (the partner may also uses immich for her/his own photos)
- [X] Take capture time from:
    - [X] JPEG files
    - [X] MP4 files
    - [X] HEIC files
    - [X] name of the file (fall back, any name containing date like Holidays_2022-07-25 21.59)
- [ ] use tags placed in exif data
- [ ] upload from remote folders
    - [ ] ssh
    - [ ] samba
    - [ ] import remote folder
- [ ] Set GPS location for images taken with a GPS-less camera based on
    - [X] Google location history
    - [ ] KML,GPX track files
- [ ] Import instead of Upload
- [x] Cleaning duplicates

# Release notes 

## Release next

### Fix #44: duplicate is not working?

At 1st run of the duplicate command, low quality images are moved to the trash and not deleted as before 1.82.
At next run, the trashed files are still seen as duplicate.
The fix consist in not considering trashed files during duplicate detection


### Fix #39: another problems with Takeout archives

I have reworked the Google takeout import to handle #39 cases. Following cases are now handled:
- normal FILE.jpg.json -> FILE.jpg
- less normal FILE.**jp**.json -> FILE.jpg
- long names truncated FIL.json -> FIL**E**.jpg
- long name with number and truncated VERY-LONG-NAM(150).json -> VERY-LONG-**NAME**(150).jpg
- duplicates names in same folder FILE.JPG(3).json -> **FILE(3)**.JPG
- edited images FILE.JSON -> FILE.JPG and **FILE-edited**.JPG

Also, there are cases where the image JSON's title is totally not related to the JSON name or the asset name.
Images are uploaded with the name found in the JSON's title field.

Thank to @bobokun for sharing details.


## Release 0.3.6
### Fix #40: Error 204 when deleting assets

## Release 0.3.5

### Fix #35: weird name cases in google photos takeout: truncated name or jp.json
Here are some weird cases found in takeout archives

example:
image title: 😀😃😄😁😆😅😂🤣🥲☺️😊😇🙂🙃😉😌😍🥰😘😗😙😚😋😛😝😜🤪🤨🧐🤓😎🥸🤩🥳😏😒😞😔😟😕🙁☹️😣😖😫😩🥺😢😭😤😠😡🤬🤯😳🥵🥶.jpg
image file: 😀😃😄😁😆😅😂🤣🥲☺️😊😇🙂🙃😉😌😍🥰😘😗😙😚😋😛.jpg
json file: 😀😃😄😁😆😅😂🤣🥲☺️😊😇🙂🙃😉😌😍🥰😘😗😙😚😋.json

example:
image title: PXL_20230809_203449253.LONG_EXPOSURE-02.ORIGINAL.jpg
image file: PXL_20230809_203449253.LONG_EXPOSURE-02.ORIGINA.jpg
json file: PXL_20230809_203449253.LONG_EXPOSURE-02.ORIGIN.json

example:
image title: 05yqt21kruxwwlhhgrwrdyb6chhwszi9bqmzu16w0 2.jpg
image file: 05yqt21kruxwwlhhgrwrdyb6chhwszi9bqmzu16w0 2.jpg
json file: 05yqt21kruxwwlhhgrwrdyb6chhwszi9bqmzu16w0 2.jp.json



### Fix #32: Albums contains album's images and all images having the same name
Once, I have a folder full of JSON files for an album, but it doesn't have any pictures. Instead, the pictures are in a folder organized by years. To fix this, I tried to match the JSON files with the pictures by their names.

The problem is that sometimes pictures have the same name in different years, so it's hard to be sure which picture goes with which JSON file. Because of this, created album contains image found in its folder, but also images having same name, taken in different years.

I decided to remove this feature. Now, if the image isn't found beside the JSON file, the JSON is ignored.


## Release 0.3.2, 0.3.3, 0.3.4
### Fix for #30 panic: time: missing Location in call to Time.In with release Windows_x86_64_0.3.1
Now handle correctly windows' timezone names even on windows.
Umm...

## Release 0.3.0 and 0.3.1
**Refactoring of Google Photo Takeout handling**

The takeout archive has flaws making the import task difficult and and error prone.
I have rewritten this part of the program to fix most of encountered error.

### google photos: can't find image of album #11 
Some image may miss from the album's folder. Those images files are located into the year folder. 
This fix looks for album images in the whole archive.

### photos with same name into the year folder #12 
Iphones and digital cameras produce images with the sequence number of 4 digits. This leads inevitably to have several images with the same number in the the year folder.

Google Photos disambiguates the files name by adding a counter at the end of the image file:
- IMG_3479.JPG
- IMG_3479(1).JPG
- IMG_3479(2).JPG

Surprisingly, matching JSON are named as 
- IMG_3479.JPG.json
- IMG_3479.JPG(1).json
- IMG_3479.JPG(2).json

This special case is now handled.

### Untitled albums are now handled correctly
Untitled albums now are named after the album's folder name.

This address partially the issue #19.

###  can't find the image with title "___", pattern: "___*.*": file does not exist: "___" #21 

The refactoring of the code don't use anymore a file pattern to find files in the archive. 
The image and the JSON file are quite identical, except for duplicate image (see #12) or when the file name is too long (how long is too long?).

Now, the program takes the image name, check if there is a JSON that matches, open it and use the title of the image to name the upload.

If the JSON isn't found, the image is uploaded with it's name in the archive, and with no date. Now all images are uploaded to immich, even when the JSON file is not found.

### MPG files not supported. #20 
Immich-go now accepts the same list of extension as the immich-server. This list is taken from the server source code.

### immich-go detects raw and jpg as duplicates #25 
The duplicate checker now uses the file name, its extension and the date of take to detect duplicates. 
So the system doesn't signal `IMG_3479.JPG` and `IMG_3479.CR2` as duplicate anymore.

### fix duplicate check before uploading #29
The date parsing now takes into account the time zone of the machine (ex: Europe/Paris). This handles correctly summer time and winter time. 
This isn't yet tested on Window or Mac machines.


## Release 0.2.3

- Improvement of duplicate command (issue#13)
  - `-yes` option to assume Yes to all questions
  - `-date` to limit the check to a a given date range
- Accept same type of files than the server (issue#15)
    - .3fr
    - .ari
    - .arw
    - .avif
    - .cap
    - .cin
    - .cr2
    - .cr3
    - .crw
    - .dcr
    - .dng
    - .erf
    - .fff
    - .gif
    - .heic
    - .heif
    - .iiq
    - .insp
    - .jpeg
    - .jpg
    - .jxl
    - .k25
    - .kdc
    - .mrw
    - .nef
    - .orf
    - .ori
    - .pef
    - .png
    - .raf
    - .raw
    - .rwl
    - .sr2
    - .srf
    - .srw
    - .tif
    - .tiff
    - .webp
    - .x3f
    - .3gp
    - .avi
    - .flv
    - .insv
    - .m2ts
    - .mkv
    - .mov
    - .mp4
    - .mpg
    - .mts
    - .webm
    - .wmv"
- new feature: add partner's assets to an album. Thanks to @mrwulf.
- fix: albums creation fails sometime

### Release 0.2.2
- improvement of date of capture when there isn't any exif data in the file
    1. test the file name for a date
    1. open the file and search for the date (.jpg, .mp4, .heic, .mov)
    1. if still not found, give the current date

> ⚠️ As of current version v1.77.0, immich fails to get the date of capture of some videos (IPhone), and place the video on the 01/01/1970.
> 
> You can use the -album to keep videos grouped in a same place despite the errors in date.

### Release 0.2.1
- Fix of -album option. uploaded images will be added into the album. Existing images will be added in the album if needed.

### Release 0.2.0
- When uploading from a directory, use the date inferred from the file name as file date.  Immich uses it as date of take. This is useful for images without Exif data.
- `duplicate` command check immich for several version of the same image, same file name, same date of capture
