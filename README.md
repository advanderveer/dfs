# dfs
Another filesystem experiment

## Possible node storage structure that could reduce random reads 

```
//low level node information, similar to a linux inode. Stored as
//
// |           Key      |       Data             |       Comment			 				 |
// 00000001						  : { ... }                #node info (a directory)
// 00000001/a.txt			  : 00000002    					 #to another node
// 00000001/b.txt			  : 00000003    					 #to another node
// 00000002						  : { ... }                #node info (a file)
// 00000002:0						: 2511E0F94...979AF0F    #chunk at file offset 0
// 00000003						  : { ... }                #node info (a file)
// 00000003:0						: 2511E0F94...979AF0F    #chunk at file offset 0 (dedup)
```
