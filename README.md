DSpace AIP Export Helper
===

This project eases exporting individual collections out of DSpace *for
development purposes* by doing the following:

- All parents of an exported collection (and their parents on up the chain) are exported
- The site is exported to get all users and groups
- Exports are stored in sorted "buckets" so they can be imported in the correct
  order to avoid trying to import something before its parent has been imported
- The site AIP's `mets.xml` is rewritten to ensure there are no empty groups
  (empty groups are replaced with users in the "Administrator" group)
- Automating the commands needed to run an export so you aren't exporting thing
  1, then thing 2, etc. manually.  Run it and walk away.

All that said, this project only gets you part of the way there with some
exports, because DSpace has some funky rules about its importing:

- Groups tied to a collection/community are not imported, even if they're
  defined in the site AIP, unless you first import the collection/community to
  which they're tied.  Which you can't do by default, because its group won't
  yet exist.  Which you can't fix with an import.  You would have to tackle
  this manually.
- Some users won't import.  I haven't figured out why.
- Your best bet is to disable group and role settings on import.  Which means
  you lose some valuable data. (see the caveats in
  https://wiki.duraspace.org/display/DSDOC6x/AIP+Backup+and+Restore#AIPBackupandRestore-SubmittinganAIPHierarchy
  for details)

And even if none of these things were true, this project has quite a few
hard-coded hacks for UO.  It won't work for other institutions most likely,
other than as an aid to seeing how we sort of almost solved the problem.

And please remember, this thing's exports are really only for development so
you can work with real-world DSpace data without exporting terabytes of data.
In a production environment, this project is garbage.
