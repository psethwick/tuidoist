initial startup

1. load cache into app lists
1. apply any ops from offline queue
1. start sync
1. on sync, update app lists, refresh from lists

local change do op:

- change local list, refresh
- add command and sync
- on sync, update app lists
