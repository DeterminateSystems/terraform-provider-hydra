#!/usr/bin/env bash
# Depends on: jq, curl, coreutils (basename, mktemp)
set -e

strFromNull() {
    local input=$1
    local expr=$2

    echo "$input" | jq -r "if $expr == null then \"\" else $expr end"
}

boolFrom() {
    local input=$1
    local expr=$2

    echo "$input" | jq -r "($expr) == true"
}

intFrom() {
    local input=$1
    local expr=$2

    echo "$input" | jq -r "$expr + 0"
}

jobsetTypeFrom() {
    local input=$1
    local expr=$2

    val=$(intFrom "$input" "$expr")
    case "$val" in
        0)
            echo "legacy";
            ;;
        1)
            echo "flake";
            ;;
        *)
            echo "UNKNOWN";
            ;;
    esac
}

jobsetStateFrom() {
    local input=$1
    local expr=$2

    val=$(intFrom "$input" "$expr")
    case "$val" in
        0)
            echo "disabled";
            ;;
        1)
            echo "enabled";
            ;;
        2)
            echo "one-shot";
            ;;
        3)
            echo "one-at-a-time";
            ;;
        *)
            echo "UNKNOWN";
            ;;
    esac
}

quotedStringFrom() {
    input=$1
    expr=$2

    value=$(echo "$input" | jq "$expr")
    if [ "$value" = "" ]; then
        value='""'
    fi

    echo "$value"
}

declarativeConfig() {
    project=$1

    file=$(echo "$project" | jq -r ".declarative.file")
    type=$(echo "$project" | jq -r ".declarative.type")
    value=$(quotedStringFrom "$project" ".declarative.value")

    if [ -n "$file" ]; then
        cat <<-TPL
  declarative {
      file  = "$file"
      type  = "$type"
      value = $value
  }
TPL
    fi
}

renderProject() {
    project=$1

    name=$(echo "$project" | jq -r ".name")
    displayname=$(strFromNull "$project" ".displayname")
    description=$(strFromNull "$project" ".description")
    homepage=$(strFromNull "$project" ".homepage")
    owner=$(strFromNull "$project" ".owner")
    enabled=$(boolFrom "$project" ".enabled")
    visible=$(boolFrom "$project" ".hidden == false")
    declarative=$(declarativeConfig "$project")

    cat <<-TPL
resource "hydra_project" "$name" {
  name         = "$name"
  display_name = "$displayname"
  homepage     = "$homepage"
  description  = "$description"
  owner        = "$owner"
  enabled      = $enabled
  visible      = $visible
TPL

    if [ -n "$declarative" ]; then
        cat <<-TPL

$declarative
TPL
    fi

    echo "}"

    echo terraform import "hydra_project.$name" "$name" >> "$importFile"
}

inputDefinitionLegacy() {
    jobset=$1

    nixexprinput=$(strFromNull "$jobset" ".nixexprinput")
    nixexprpath=$(strFromNull "$jobset" ".nixexprpath")

    cat <<-TPL
  nix_expression {
    file  = "$nixexprpath"
    input = "$nixexprinput"
  }

TPL

    while read -r input; do
        name=$(strFromNull "$input" ".name")
        type=$(strFromNull "$input" ".type")
        value=$(quotedStringFrom "$input" ".value")
        notify_committers=$(boolFrom "$input" ".emailresponsible")

    cat <<-TPL
  input {
    name              = "$name"
    type              = "$type"
    value             = $value
    notify_committers = $notify_committers
  }

TPL

    done < <(echo "$jobset" | jq -c ".inputs | to_entries | sort_by(.key) | .[] | .value")
}

inputDefinitionFlake() {
    jobset=$1

    flake=$(strFromNull "$jobset" ".flake")

    cat <<-TPL
  flake_uri = "$flake"

TPL
}

renderJobset() {
    project=$1
    name=$2
    jobset=$3

    state=$(jobsetStateFrom "$jobset" ".enabled")
    description=$(strFromNull "$jobset" ".description")
    type=$(jobsetTypeFrom "$jobset" ".type")
    visible=$(boolFrom "$jobset" ".visible")
    keep_evaluations=$(intFrom "$jobset" ".keepnr")
    scheduling_shares=$(intFrom "$jobset" ".schedulingshares")
    check_interval=$(intFrom "$jobset" ".checkinterval")

    email_notifications=$(boolFrom "$jobset" ".enableemail")
    email_override=$(strFromNull "$jobset" ".emailoverride")

    case "$type" in
        "legacy")
            inputdefinition=$(inputDefinitionLegacy "$jobset")
            ;;
        "flake")
            inputdefinition=$(inputDefinitionFlake "$jobset")
            ;;
        *)
            inputdefinition="UNKNOWN INPUT TYPE"
            ;;
    esac

    resourcename=$(echo "${project}_$name" | tr '.' '_')

    cat <<-TPL
resource "hydra_jobset" "$resourcename" {
  project     = hydra_project.$project.name
  state       = "$state"
  visible     = $visible
  name        = "$name"
  type        = "$type"
  description = "$description"

$inputdefinition

  check_interval    = $check_interval
  scheduling_shares = $scheduling_shares
  keep_evaluations  = $keep_evaluations

  email_notifications = $email_notifications
  email_override      = "$email_override"
}
TPL

    echo terraform import "hydra_jobset.$resourcename" "${project}/$name" >> "$importFile"
}

generate() {
    i=0

    while read -r project; do
        projectname=$(echo "$project" | jq -r .name)

        if [ ! -e "$generatedDir/generated.$projectname.tf" ] || [ -n "${FORCE_IMPORT_ALL-}" ]; then
            (( i+=1 ))

            (
                (
                    echo "Processing project '$projectname'..." >&2
                    renderProject "$project"
                    declfile=$(echo "$project" | jq -r .declarative.file)

                    if [ -z "$declfile" ]; then
                        while read -r jobsetname; do
                            if [ "$jobsetname" = ".jobsets" ]; then
                                echo >&2
                                echo "WARNING: $project is not declarative, but has a .jobsets jobset which has been ignored." >&2
                                echo "         See https://github.com/NixOS/hydra/issues/960 for more information." >&2
                                echo >&2
                                continue
                            fi

                            echo "Processing jobset '$projectname/$jobsetname'..." >&2
                            echo
                            jobset="$(curl --silent --header "Accept: application/json" "$serverRoot/jobset/$projectname/$jobsetname" | jq .)"
                            renderJobset "$projectname" "$jobsetname" "$jobset"
                        done < <(echo "$project" | jq -r ".jobsets | sort | .[]")
                    fi
                ) > "$generatedDir/generated.$projectname.tf"
            ) &
        fi

    done < <(jq -c ".[]" < "$inputFile")

    wait

    echo "Done, generated Terraform configuration for $i projects."
    [ -e "$importFile" ] && echo "The generated 'terraform import' statements have been saved to '$importFile'."
}

help() {
    echo "Usage: $(basename "$0") <server-root> <out-dir> <import-file>"
    echo
    echo "    Arguments:"
    echo "        <server-root>    The root of the Hydra server to import projects and jobsets from."
    echo "        <out-dir>        The directory to output generated Terraform configuration files to."
    echo "        <import-file>    Where to write the generated list of 'terraform import' statements."

    exit 1
}

main() {
    serverRoot="$1"
    generatedDir="$2"
    importFile="$3"

    if [ -z "$serverRoot" ] || [ -z "$generatedDir" ] || [ -z "$importFile" ]; then
        help
    fi

    if [ ! -d "$generatedDir" ]; then
        echo "Output directory '$generatedDir' is not a directory or does not exist."
        exit 1
    fi

    if [ -e "$importFile" ]; then
        echo "Import file '$importFile' already exists."
        exit 1
    fi

    inputFile=$(mktemp -t projects.json.XXXXXXXXXX)
    finish() {
        rm -f "$inputFile"
    }
    trap finish EXIT

    echo "Fetching projects.json..."
    curl --silent --show-error --header "Accept: application/json" "$serverRoot" > "$inputFile"

    echo "Starting generation..."
    generate
}

main "$@"
