### Format of the Commit Message

The subject line should feature **what**.

The body of the commit should describe **why**. Use correct punctuation (capital letters, etc. ...). 

If the commit contains a fix to a recorded issue, then reference the issue ID in the footer.

```
Add a method to ping Mongo

To have the ability to check if Mongo is healthy or not.

Fixes #23
```

The format can be described more formally as follows:

```
<what the change is>
<BLANK LINE>
<why this change was made>
<BLANK LINE>
<footer>
```

The first line is the subject and should be no longer than 50 characters, the
second line is always blank, and other lines should be wrapped at 70 characters.
This allows the message to be easier to read on GitHub as well as in various
git tools.


As commit messages are meant to be kept compact, comprehensive descriptions and explanations may be written for Pull Requests in the description.  

#### DO
Write the summary line and description of what you have done in the imperative mode, that is as if you were commanding someone; start the line with "Fix", "Add", "Change" instead of "Fixed", "Added", "Changed".

Always leave the second line blank.

Line break the commit message (to make the commit message readable without having to scroll horizontally).

#### DON'T 
End the summary line with a period - it's a title and titles don't end with a period.