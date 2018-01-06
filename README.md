gm\_txt
---

Edit GameMaker objects as plaintext files.

```
Sprite sprPlayer
Persistent
Depth -10

---Create
moveSpeed = 6

---Step
if keyboard_check(vk_right) {
    x += moveSpeed
}

---Collision objStar
with other instance_destroy()
```

A directory gm\_txt is created and monitored for changes.

```
bread.gmx/
    bread.project.gmx
    gm_txt/
        objBread.gmo
        scrButter.gml
```

[Download (TODO)]()
---

Usage
---
```
gm_txt.exe [project_file]
```
If you provide no arguments, it'll open a file picker dialog (thanks to [sqweek/dialog](https://github.com/sqweek/dialog)).

Type "help" for a cheat sheet of event names.

Building
---

```
go install
```
:)
