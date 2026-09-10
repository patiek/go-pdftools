/*
Package pdftk provides wrapper functions for calling PDFtk commands.

Expects command line executable of pdftk or pdftk-java to be installed.

Commands that take an io.Reader do not return while that reader is blocked in
Read, even after ctx is done, so pass an *os.File or in-memory reader when the
call must be bounded.

By: Patrick Brown
*/
package pdftk
