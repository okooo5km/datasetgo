# DatasetGo

datasetgo 是一款用于处理深度学习目标检测数据集的命令行小工具。目前工具支持三种格式 COCO、PascalVOC 和 CreateML。

## RoadMap

datasetgo 将具备以下子命令。

- [x] convert: 转换数据集格式子命令；
- [ ] split: 分配数据集到训练集、测试集、验证集；
- [ ] list: 列出数据集的基本信息；
- [ ] analyse： 分析数据集特征；

## Usage

介绍各子命令的基本使用。

### convert 子命令

```shell
> datasetgo convert -h
A subcommand to convert the dataset format. The supported
formats as follows:
- coco: COCO
- voc: PascalVOC
- createml: Create ML(apple)

Usage:
  datasetgo convert [flags] dataset-path

Flags:
  -h, --help                   help for convert
  -i, --input-format string    the format of the source dataset
  -o, --output-format string   the format of the outputed dataset
  -p, --output-path string     the path of the outputed dataset, a file or directory

Global Flags:
  -v, --verbose   verbose output
```

比如将 COCO 数据集转换成 PascalVOC 数据集（转换的数据集文件自动导出到 coco json 的目录下），使用命令：

```shell
datasetgo convert -i coco -o voc the/dataset/path/of/coco/json/file.json
```

### split 子命令

`待添加`

### list 子命令

`待添加`

### analyse 子命令

`待添加`
