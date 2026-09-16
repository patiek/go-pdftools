/*
Package pdftk provides wrapper functions for calling PDFtk commands.

Expects command line executable of pdftk or pdftk-java to be installed.

Inputs are io.Readers, read at most once and never closed. An unread regular
*os.File is passed to pdftk by name and left unread. Any other reader is
copied to a temp file first (see OptionTempDir) because pdftk needs seekable
files and would hold a whole stdin input in memory. Copying stops between
Reads once ctx is done but cannot interrupt a blocked Read.

By: Patrick Brown
*/
package pdftk
