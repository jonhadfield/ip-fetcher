package main

const (
	errStdoutOrPathRequired = "error: must specify at least one of stdout and Path"
	errAggregateNeedsLines  = "error: --aggregate requires --lines (or --format lines)"
	usageWriteToStdout      = "write to stdout"
	usageWhereToSaveFile    = "where to save the file"
	usageLinesOutput        = "output newline separated ip prefixes"
	usageAggregate          = "reduce --lines output: exact (no holes) or cover (may fill holes)"
	fmtDataWrittenTo        = "Data written to %s\n"

	flagPath      = "Path"
	flagStdout    = "stdout"
	flagFormat    = "format"
	flagIPv4      = "ipv4"
	flagAggregate = "aggregate"

	formatLines = "lines"
	formatJSON  = "json"
	formatCSV   = "csv"
	formatYAML  = "yaml"
)
