package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/valerii/hub/backend/internal/schemamaintenance/rootfk"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "schema-root-fk:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("schema-root-fk", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	expected := flags.String("expected", "", "frozen six-field manifest")
	desired := flags.String("desired", "", "Atlas desired-state inspect SQL")
	replay := flags.String("replay", "", "Atlas full-replay inspect SQL")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 || *expected == "" || *desired == "" || *replay == "" {
		return errors.New("usage: schema-root-fk --expected <manifest> --desired <atlas-inspect.sql> --replay <atlas-inspect.sql>")
	}
	evidence, err := rootfk.VerifyFiles(*expected, *desired, *replay)
	if err != nil { return err }
	fmt.Println(evidence.String())
	return nil
}
