# gogit

自动执行 pull & push 的工具，解放双手。🙌

## install

根目录执行：

```sh
$ go build
```

## how to use

```
$ ./gogit -h
Usage of ./gogit:
  -c int
    	执行次数（值 <=0 不限制次数，默认不限制次数）
  -i int
    	执行间隔时间，单位：秒 (default 5)
  -pull
    	指定执行 pull 操作
```

```sh
## 指定执行间隔时间为 10 秒
$ ./gogit -i 10
```



