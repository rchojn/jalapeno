Feature: Execute recipes
  Executing Jalapeno recipes to template out projects

  Scenario: Execute single recipe
    Given a project directory
    And a recipes directory
    And a recipe "foo" that generates file "README.md"
    When I execute recipe "foo"
    Then execution of the recipe has succeeded
    And the project directory should contain file "README.md"
    And the sauce file contains a sauce in index 0 which should have property "recipe.name" with value "^foo$"
    And the sauce file contains a sauce in index 0 which has a valid ID

  Scenario: Execute single recipe from remote registry
    Given a project directory
    And a recipes directory
    And a recipe "foo" that generates file "README.md"
    And a local OCI registry
    And the recipe "foo" is pushed to the local OCI repository "foo:v0.0.1"
    When I execute the recipe from the local OCI repository "foo:v0.0.1"
    Then execution of the recipe has succeeded
    And no errors were printed
    And the project directory should contain file "README.md"

  Scenario: Execute multiple recipes
    Given a project directory
    And a recipes directory
    And a recipe "foo" that generates file "README.md"
    And a recipe "bar" that generates file "Taskfile.yml"
    When I execute recipe "foo"
    Then execution of the recipe has succeeded
    When I execute recipe "bar"
    Then execution of the recipe has succeeded
    And no errors were printed
    And the project directory should contain file "README.md"
    And the project directory should contain file "Taskfile.yml"
    And the sauce file contains a sauce in index 0 which should have property "recipe.name" with value "^foo$"
    And the sauce file contains a sauce in index 1 which should have property "recipe.name" with value "^bar$"

  Scenario: New recipe conflicts with the previous recipe
    Given a project directory
    And a recipes directory
    And a recipe "foo" that generates file "README.md"
    And a recipe "bar" that generates file "Taskfile.yml"
    And a recipe "quux" that generates file "Taskfile.yml"
    When I execute recipe "foo"
    And no errors were printed
    Then execution of the recipe has succeeded
    When I execute recipe "bar"
    And no errors were printed
    Then execution of the recipe has succeeded
    When I execute recipe "quux"
    Then CLI produced an error "file 'Taskfile.yml' was already created by recipe 'bar'"