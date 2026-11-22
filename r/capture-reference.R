# capture-reference.R: generates the committed samples and records R's trimmed means (spike 001).
#
# Run from the repository root on the Mac: Rscript r/capture-reference.R
# It writes:
#   data/<name>.csv                       one value per line under the header "value"
#   testdata/r-reference/trimmed-means.tsv R's results for every sample and trimming case
#   testdata/r-reference/semantics.tsv     how R behaves at the edges (negative trim, NA, trim >= 0.5, ...)
#   testdata/r-reference/r-session.txt     R version and the settings the results depend on
# It reads no personal data and writes no user name, host name or home path.

seed <- 9431

# Values are written with 17 significant digits, enough for every float64 to survive the round
# trip to text and back exactly (constitution 8.3).
format_value <- function(values) sprintf("%.17g", values)

# --- Samples -------------------------------------------------------------------------------------

# Integers: values around 50 with 5 % large outliers, so that trimming visibly matters.
make_integers <- function(n) {
  values <- as.integer(round(rnorm(n, mean = 50, sd = 10)))
  outliers <- max(1L, n %/% 20L)
  values[sample.int(n, outliers)] <- as.integer(round(runif(outliers, 500, 1000)))
  values
}

# Floats: a 90/10 mixture of N(0, 1) and N(0, 10), the heavy-tailed shape the course's robust
# estimation example uses.
make_floats <- function(n) {
  heavy <- runif(n) < 0.10
  narrow <- rnorm(n, 0, 1)
  wide <- rnorm(n, 0, 10)
  ifelse(heavy, wide, narrow)
}

# Sizes: 100 (the assignment's minimum), 101 (n * 0.05 is not a whole number), 1000.
set.seed(seed)
samples <- list(
  "integers-100"  = make_integers(100L),
  "integers-101"  = make_integers(101L),
  "integers-1000" = make_integers(1000L),
  "floats-100"    = make_floats(100L),
  "floats-101"    = make_floats(101L),
  "floats-1000"   = make_floats(1000L)
)

dir.create("data", showWarnings = FALSE)
dir.create("testdata/r-reference", recursive = TRUE, showWarnings = FALSE)

# write_sample saves a sample and reads it back, so that every result below is computed from the
# exact bytes Go will read, not from R's memory.
write_sample <- function(name, values) {
  path <- file.path("data", paste0(name, ".csv"))
  text <- if (is.integer(values)) as.character(values) else format_value(values)
  writeLines(c("value", text), path)
  read_back <- readLines(path)[-1]
  values_read <- if (is.integer(values)) as.integer(read_back) else as.numeric(read_back)
  stopifnot(identical(values_read, values))
  values_read
}
samples <- Map(write_sample, names(samples), samples)

# --- Trimmed means -------------------------------------------------------------------------------

# Our asymmetric definition, evaluated in R: sort, drop floor(n * low) values from the bottom and
# floor(n * high) from the top, average what is left. R has no built-in for this.
defined_trimmed_mean <- function(values, low, high) {
  n <- length(values)
  drop_low <- floor(n * low)
  drop_high <- floor(n * high)
  mean(sort(values)[(drop_low + 1):(n - drop_high)])
}

symmetric_trims <- c(0, 0.05, 0.1, 0.2, 0.25, 0.29, 0.45, 0.49)
asymmetric_trims <- list(c(0, 0.05), c(0.05, 0), c(0.1, 0.2), c(0.2, 0.1), c(0.05, 0.25), c(0.29, 0.01))

rows <- list()
add_row <- function(...) rows[[length(rows) + 1]] <<- data.frame(..., stringsAsFactors = FALSE)

for (name in names(samples)) {
  values <- samples[[name]]
  n <- length(values)
  kind <- if (is.integer(values)) "integer" else "double"
  for (trim in symmetric_trims) {
    builtin <- mean(values, trim = trim)
    full_sort <- defined_trimmed_mean(values, trim, trim)
    as_double <- mean(as.double(values), trim = trim)
    add_row(sample = name, kind = kind, n = n, low = trim, high = trim, method = "mean(x, trim)",
            value = format_value(builtin), dropped_low = floor(n * trim), dropped_high = floor(n * trim),
            identical_to_full_sort = identical(builtin, full_sort),
            identical_as_double = identical(builtin, as_double),
            full_sort_value = format_value(full_sort))
  }
  for (pair in asymmetric_trims) {
    defined <- defined_trimmed_mean(values, pair[1], pair[2])
    add_row(sample = name, kind = kind, n = n, low = pair[1], high = pair[2], method = "defined",
            value = format_value(defined), dropped_low = floor(n * pair[1]),
            dropped_high = floor(n * pair[2]), identical_to_full_sort = NA, identical_as_double = NA,
            full_sort_value = format_value(defined))
  }
}
write.table(do.call(rbind, rows), "testdata/r-reference/trimmed-means.tsv",
            sep = "\t", quote = FALSE, row.names = FALSE)

# --- Edge behaviour ------------------------------------------------------------------------------

# probe records what R returns for one expression, or the warning or error it raises.
probe <- function(expression_text) {
  result <- tryCatch({
    value <- eval(parse(text = expression_text))
    paste(format_value(value), collapse = " ")
  },
  warning = function(w) paste("WARNING:", conditionMessage(w)),
  error = function(e) paste("ERROR:", conditionMessage(e)))
  data.frame(expression = expression_text, result = result, stringsAsFactors = FALSE)
}
edge_cases <- c(
  "floor(100 * 0.29)", "floor(100 * 0.07)", "floor(101 * 0.05)", "floor(57 * 0.1)",
  "mean(c(1, 2, 3, 4, 100), trim = -0.1)",
  "mean(c(1, 2, 3, 4, 100), trim = 0.5)", "mean(c(1, 2, 3, 4, 100), trim = 0.7)",
  "median(c(1, 2, 3, 4, 100))", "mean(c(1, 2, 3, 4, 100, 7), trim = 0.5)",
  "mean(c(2, 4), trim = 0.49)", "mean(1, trim = 0.4)", "mean(c(1L, 2L, 3L), trim = 0.1)",
  "mean(1:10, trim = 0.1)", "mean(c(1, NA, 3), trim = 0.1)", "mean(c(1, NaN, 3), trim = 0.1)",
  "mean(numeric(0), trim = 0.1)", "mean(numeric(0))", "mean(c(1, 2, 3), trim = c(0.1, 0.2))",
  "mean(c(1, 2, 3), trim = NA)", "mean(c(-Inf, 1:8, Inf), trim = 0.1)", "mean(c(-Inf, 1:8, Inf))",
  "mean(c(1e308, 1e308))", "mean(c(1e308, 1e308, 1, 1), trim = 0.25)", "mean(\"a\", trim = 0.1)"
)
write.table(do.call(rbind, lapply(edge_cases, probe)), "testdata/r-reference/semantics.tsv",
            sep = "\t", quote = FALSE, row.names = FALSE)

# --- Session -------------------------------------------------------------------------------------

writeLines(c(
  paste("R:", R.version.string),
  paste("platform:", R.version$platform),
  paste("RNGkind:", paste(RNGkind(), collapse = " / ")),
  paste("seed:", seed),
  paste("sizeof long double:", .Machine$sizeof.longdouble),
  paste("long.double capability:", capabilities("long.double")),
  paste("captured:", format(Sys.time(), "%Y-%m-%d %H:%M %Z"))
), "testdata/r-reference/r-session.txt")

cat("capture-reference: wrote", length(samples), "samples,", length(rows), "trimmed means,",
    length(edge_cases), "edge cases\n")
