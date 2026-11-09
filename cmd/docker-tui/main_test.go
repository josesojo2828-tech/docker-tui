package main

import (
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		nombre   string
		args     []string
		wantMode string
		wantRest []string
		wantErr  bool
	}{
		{
			nombre:   "sin argumentos usa all",
			args:     nil,
			wantMode: "all",
		},
		{
			nombre:   "flags sin subcomando siguen funcionando",
			args:     []string{"-port", "9000"},
			wantMode: "all",
			wantRest: []string{"-port", "9000"},
		},
		{
			nombre:   "subcomando tui",
			args:     []string{"tui"},
			wantMode: "tui",
			wantRest: []string{},
		},
		{
			nombre:   "subcomando web con flags",
			args:     []string{"web", "-port", "9001"},
			wantMode: "web",
			wantRest: []string{"-port", "9001"},
		},
		{
			nombre:   "subcomando all con flags",
			args:     []string{"all", "-no-web"},
			wantMode: "all",
			wantRest: []string{"-no-web"},
		},
		{
			nombre:  "comando invalido devuelve error",
			args:    []string{"foo"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.nombre, func(t *testing.T) {
			mode, rest, err := parseArgs(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Fatal("se esperaba un error y no hubo ninguno")
				}
				return
			}

			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			if mode != tt.wantMode {
				t.Errorf("mode = %q, se esperaba %q", mode, tt.wantMode)
			}
			if !reflect.DeepEqual(rest, tt.wantRest) {
				t.Errorf("rest = %#v, se esperaba %#v", rest, tt.wantRest)
			}
		})
	}
}
