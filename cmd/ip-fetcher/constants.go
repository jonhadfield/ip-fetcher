package main

const (
	errStdoutOrPathRequired = "error: must specify at least one of stdout and Path"
	usageWriteToStdout      = "write to stdout"
	usageWhereToSaveFile    = "where to save the file"
	usageLinesOutput        = "output newline separated ip prefixes"
	fmtDataWrittenTo        = "Data written to %s\n"

	flagPath   = "Path"
	flagStdout = "stdout"
	flagFormat = "format"
	flagIPv4   = "ipv4"

	formatLines = "lines"
	formatJSON  = "json"
	formatCSV   = "csv"
	formatYAML  = "yaml"
)
