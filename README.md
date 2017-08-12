NiceObjects
-----------

Seamlessly edit GameMaker objects in a friendly text format (along with scripts).

```
Sprite sprPlayer
Persistent
Depth -10

---Create
moveSpeed = 6

---Step
if keyboard_check(vk_right) {
    hspeed = moveSpeed
} else {
    hspeed = 0
}

---Collision objStar
with other instance_destroy()
```

NiceObjects creates and monitors a directory of translated GameMaker object
files in the above format (.ogml). When these files are modified or created,
they are translated back to the actual project's files. Script files (.gml) are
simply copied back and forth since they require no translation.

Sample session output:

```
Initializing NiceObjects directory...
Listening (Ctrl-C to quit) (type "help" for help)
[21:42:16] Translated objStar
[21:42:24] Error reading human object file C:\Patrick\Projects\Go\src\github.com\patrickgh3\NiceObje
cts\NiceObjects\objPlayer.ogml: Unrecognized event title Createe on line 5
[21:42:27] Translated objPlayer
Quit signal received, removing NiceObjects directory...
Success
```

The lists of object properties and event names, along with general documentation
are found by typing "help" after starting the tool. Also see readme_dist.txt for
helpful GML syntax links.

I made this to satisfy my own need, and I hope it's useful to a few other people
too. Don't hesitate to contact me if you have any suggestions or bug reports.
Discord: Patrick#0303

Uses [sqweek/dialog](https://github.com/sqweek/dialog) for the file open dialog.

For Vim syntax highlighting check out [nessss/vim-gml](https://github.com/nessss/vim-gml)
(you'll want to add .ogml to the ftdetect file)

[Download (TODO)]()
---

Contributing
------------
Improvements are welcome! Just don't wildly increase the scope of the tool.

