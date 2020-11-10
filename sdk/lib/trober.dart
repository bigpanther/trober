import 'package:openapi_generator_annotations/openapi_generator_annotations.dart';

@Openapi(
    additionalProperties: AdditionalProperties(
        pubName: 'trober_sdk',
        pubAuthor: 'Big Panther Inc.',
        pubVersion: '0.0.1',
        pubAuthorEmail: 'info@bigpanther.ca',
        pubDescription: 'The trober SDK',
        pubHomepage: 'https://bigpanther.ca'),
    inputSpecFile: 'lib/trober.yaml',
    apiPackage: 'trober_sdk',
    generatorName: Generator.DART2_API,
    alwaysRun: true,
    fetchDependencies: false,
    outputDirectory: 'trober_sdk')
class Trober extends OpenapiGeneratorConfig {}
