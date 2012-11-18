// Support the Gherkin language, as found in Ruby's Cucumber and Python's Lettuce projects.
package gherkin

import (
    re "regexp"
    "strings"
)

type stepdef struct {
    r *re.Regexp
    f func()
}

func createstep(p string, f func()) stepdef {
    r, _ := re.Compile(p)
    return stepdef{r, f}
}

func (s stepdef) execute(line string) bool {
    if s.r.MatchString(line) {
        s.f()
        return true
    }
    return false
}

type Runner struct {
    steps []stepdef
    StepCount int
    scenarioIsPending bool
    background []string
    collectBackground bool
    setUp func()
    tearDown func()
}

// Register a set-up function to be called at the beginning of each scenario
func (r *Runner) SetSetUpFn(setUp func()) {
    r.setUp = setUp
}

// Register a tear-down function to be called at the end of each scenario
func (r *Runner) SetTearDownFn(tearDown func()) {
    r.tearDown = tearDown
}

// The recommended way to create a gherkin.Runner object.
func CreateRunner() *Runner {
    return &Runner{make([]stepdef, 1), 0, false, make([]string, 0), false, nil, nil}
}

// Register a step definition. This requires a regular expression
// pattern and a function to execute.
func (r *Runner) Register(pattern string, f func()) {
    r.steps = append(r.steps, createstep(pattern, f))
}

func (r *Runner) executeFirstMatchingStep(line string) {
    for _, step := range r.steps {
        if step.execute(line) {
            r.StepCount++
            return
        }
    }
}

func (r *Runner) callSetUp() {
    if r.setUp != nil {
        r.setUp()
    }
}

func (r *Runner) callTearDown() {
    if r.tearDown != nil {
        r.tearDown()
    }
}

func (r *Runner) parseAsStep(line string) (bool, string) {
    givenMatch, _ := re.Compile(`(Given|When|Then|And|But|\*) (.*?)\s*$`)
    if s := givenMatch.FindStringSubmatch(line); !r.scenarioIsPending && s != nil && len(s) > 1 {
        return true, s[2]
    }
    return false, ""
}

func (r *Runner) isScenarioLine(line string) (bool) {
    scenarioMatch, _ := re.Compile(`Scenario:\s*(.*?)\s*$`)
    if s := scenarioMatch.FindStringSubmatch(line); s != nil {
        return true
    }
    return false
}

func (r *Runner) isFeatureLine(line string) bool {
    featureMatch, _ := re.Compile(`Feature:\s*(.*?)\s*$`)
    if s := featureMatch.FindStringSubmatch(line); s != nil {
        return true
    }
    return false
}

func (r *Runner) isBackgroundLine(line string) bool {
    backgroundMatch, _ := re.Compile(`Background:`)
    if s := backgroundMatch.FindStringSubmatch(line); s != nil {
        return true
    }
    return false
}

func (r *Runner) executeStep(line string) {
    if !r.collectBackground {
        r.executeFirstMatchingStep(line)
    } else {
        r.background = append(r.background, line)
    }
}

func (r *Runner) startScenario() {
    r.callTearDown()
    r.collectBackground = false
    r.scenarioIsPending = false
    r.callSetUp()
    for _, bline := range r.background {
        r.executeFirstMatchingStep(bline)
    }
}

func (r *Runner) step(line string) {
    defer func() {
        if rec := recover(); rec != nil {
            if rec == "Pending" {
                r.scenarioIsPending = true
            } else {
                panic(rec)
            }
        }
    }()

    if isStep, data := r.parseAsStep(line); isStep { 
        r.executeStep(data)
    } else if r.isScenarioLine(line) {
        r.startScenario()
    } else if r.isFeatureLine(line) { 
        // Do Nothing!
    } else if r.isBackgroundLine(line) {
        r.collectBackground = true
    }
}

// Once the step definitions are Register()'d, use Execute() to
// parse and execute Gherkin data.
func (r *Runner) Execute(file string) {
    lines := strings.Split(file, "\n")
    for _, line := range lines {
        r.step(line)
    }
    r.callTearDown()
}

// Use this function to let the user know that this
// test is not complete.
func Pending() {
    panic("Pending")
}
