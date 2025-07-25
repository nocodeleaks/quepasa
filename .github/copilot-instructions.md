## Common Guidelines
* code comments should always be in English;
* response to user queries should be in IDE current language;
* avoid to change code that was not related to the query;
* when agent has to change a method and it change the async status, the agent should update the method callers too;
* for extensions methods use always "source" as default parameter name
* use one file for each class
* for #region tags: no blank lines between consecutive regions, but always add one blank line after region opening and one blank line before region closing
* do not try to build if you just changed the code comments or documentation files;
* **when making relevant code changes, always create or update internal documentation following the Internal Documentation Guidelines**;
* sempre que for criar um método de extensão, utilize 'source' como nome de parametro para o objeto extendido