Run 10 ue:
```bash
sudo ./mssim multi-ue -n 10
```

Setting 10 ue with wait start interval (500ms)
- note: interval should than 100
- default: interval = 500
```bash
sudo ./mmsim multi-ue -n 10 -tr 500
```

Write to log (``send pdu session establishment -> response accept``)
```bash
sudo ./mssim multi-ue -lg out.log
```

! `Maybe you need: Setting buffer space available` - read file [run](./run.sh)
