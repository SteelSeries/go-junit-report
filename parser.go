package main

import (
    "bufio"
    "io"
    "regexp"
    "strconv"
    "strings"
)

type Result int

const (
    PASS Result = iota
    FAIL
)

type Report struct {
    Packages []Package
}

type Package struct {
    Name      string
    Time      int
    Tests     []Test
    TestCount int
    FailCount int
}

type Test struct {
    Name   string
    Time   int
    Result Result
    Output []string
}

var (
    regexStatus = regexp.MustCompile(`^--- (PASS|FAIL): (.+) \((\d+\.\d+) seconds\)$`)
    regexPassed = regexp.MustCompile(`^OK: (\d+) passed$`)
    regexFailed = regexp.MustCompile(`^OOPS: (\d+) passed, (\d+) FAILED$`)
    regexResult = regexp.MustCompile(`^(ok|FAIL)\s+(.+)\s+(\d+\.\d+)s$`)
)

func Parse(r io.Reader) (*Report, error) {
    reader := bufio.NewReader(r)

    report := &Report{make([]Package, 0)}

    // keep track of tests we find
    tests := make([]Test, 0)

    // current test
    var test *Test

    // parse lines
    for {
        l, _, err := reader.ReadLine()
        if err != nil && err == io.EOF {
            break
        } else if err != nil {
            return nil, err
        }

        line := string(l)
        println("processing:", line)
        if strings.HasPrefix(line, "=== RUN ") {
            // start of a new test
            if test != nil {
                tests = append(tests, *test)
            }

            test = &Test{
                Name:   line[8:],
                Result: FAIL,
                Output: make([]string, 0),
            }
        } else if matches := regexPassed.FindStringSubmatch(line); len(matches) == 2 {
            println("matched pass string")
            // all tests in this package are finished
            if test != nil {
                tests = append(tests, *test)
                test = nil
            }

            testCount, _ := strconv.Atoi(matches[1])
            report.Packages = append(report.Packages, Package{
                Tests:     tests,
                TestCount: testCount,
                FailCount: 0,
            })

            tests = make([]Test, 0)
        } else if matches := regexFailed.FindStringSubmatch(line); len(matches) == 3 {
            println("matched fail string")
            // all tests in this package are finished
            if test != nil {
                tests = append(tests, *test)
                test = nil
            }

            passCount, _ := strconv.Atoi(matches[1])
            failCount, _ := strconv.Atoi(matches[2])
            report.Packages = append(report.Packages, Package{
                Tests:     tests,
                TestCount: passCount + failCount,
                FailCount: failCount,
            })

            tests = make([]Test, 0)
        } else if matches := regexResult.FindStringSubmatch(line); len(matches) == 4 {
            println("matched regexResult")
            // all tests in this package are finished
            if test != nil {
                tests = append(tests, *test)
                test = nil
            }

            println("setting package name to", matches[2])

            pkg := report.Packages[len(report.Packages)-1]
            pkg.Name = matches[2]
            pkg.Time = parseTime(matches[3])
            pkg.Tests = tests

            tests = make([]Test, 0)
        } else if test != nil {
            if matches := regexStatus.FindStringSubmatch(line); len(matches) == 4 {
                // test status
                if matches[1] == "PASS" {
                    test.Result = PASS
                } else {
                    test.Result = FAIL
                }

                test.Name = matches[2]
                test.Time = parseTime(matches[3]) * 10

            } else if strings.HasPrefix(line, "\t") {
                // test output
                test.Output = append(test.Output, line[1:])
            }
        }
    }

    return report, nil
}

func parseTime(time string) int {
    t, err := strconv.Atoi(strings.Replace(time, ".", "", -1))
    if err != nil {
        return 0
    }
    return t
}
