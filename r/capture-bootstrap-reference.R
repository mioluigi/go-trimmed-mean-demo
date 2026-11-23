# capture-bootstrap-reference.R: R's reference for spec 002 (the bootstrap study). Run from the repository
# root on the Mac, through scripts/capture-bootstrap-r.sh:
#   Rscript r/capture-bootstrap-reference.R /path/to/run-bootstrap-robust.R
#
# It runs the course file UNCHANGED (it is course material and is not committed), and saves, in
# testdata/bootstrap-reference/:
#   rnorm-seed-9999.txt     5,000 rnorm() draws after set.seed(9999), for Go's generator test
#   mixture-seed-9999.txt   2,000 draws made the way the course's rmixnorm() makes them, after set.seed(9999)
#   course-listing.txt      the course file's own printed output
#   study-results.tsv       every row of the course's studyResults, at full precision
#   standard-errors.tsv     the course's standard errors per n, at full precision
#   r-session.txt           R version, generators, B
# Values are written with sprintf("%.17g"), which is exact.

args <- commandArgs(trailingOnly = TRUE)
if (length(args) != 1 || !file.exists(args[1])) {
  stop("give the path to the course file run-bootstrap-robust.R as the only argument")
}
out <- "testdata/bootstrap-reference"
dir.create(out, recursive = TRUE, showWarnings = FALSE)
format_value <- function(values) sprintf("%.17g", values)

set.seed(9999)
writeLines(c("value", format_value(rnorm(5000))), file.path(out, "rnorm-seed-9999.txt"))

set.seed(9999)
mixture <- vapply(1:2000, function(i) {
  if (runif(1) >= 0.01) rnorm(1, 100, 10) else rnorm(1, 200, 1)
}, numeric(1))
writeLines(c("value", format_value(mixture)), file.path(out, "mixture-seed-9999.txt"))

# The course file, unchanged, in its own environment; its printed lines are kept.
study <- new.env()
listing <- capture.output(source(args[1], local = study))
writeLines(listing, file.path(out, "course-listing.txt"))

results <- study$studyResults
write.table(data.frame(n = results$n, sampleMean = format_value(results$sampleMean),
                       bootMean = format_value(results$bootMean), bootMedian = format_value(results$bootMedian),
                       bootTrimmedMean = format_value(results$bootTrimmedMean)),
            file.path(out, "study-results.tsv"), sep = "\t", quote = FALSE, row.names = FALSE)

standard_errors <- do.call(rbind, lapply(sort(unique(results$n)), function(n) {
  rows <- results[results$n == n, ]
  data.frame(n = n, clt = format_value(study$mainPopSD / sqrt(n)), sampleMean = format_value(sd(rows$sampleMean)),
             bootMean = format_value(sd(rows$bootMean)), bootMedian = format_value(sd(rows$bootMedian)),
             bootTrimmedMean = format_value(sd(rows$bootTrimmedMean)))
}))
write.table(standard_errors, file.path(out, "standard-errors.tsv"), sep = "\t", quote = FALSE, row.names = FALSE)

writeLines(c(
  paste("R:", R.version.string),
  paste("platform:", R.version$platform),
  paste("RNGkind:", paste(RNGkind(), collapse = " / ")),
  paste("B:", study$B),
  paste("iterations rows:", nrow(results)),
  paste("captured:", format(Sys.time(), "%Y-%m-%d %H:%M %Z"))
), file.path(out, "r-session.txt"))
cat("capture-bootstrap-reference: wrote", nrow(results), "study rows and", nrow(standard_errors), "standard-error rows\n")
