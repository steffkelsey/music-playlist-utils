import typer
from typing_extensions import Annotated


def pymsctl(
    input_file: Annotated[str, typer.Option(help="The json report with albums to search for")],
    output_dir: Annotated[str, typer.Option(help="The directory to size downloaded msuic files")],
    dry_run: Annotated[bool, typer.Option(help="Dry-run to print json results to stdout")] = False
):
    result = f"input-file: {input_file}\n"
    result += f"output-dir: {output_dir}\n"
    result += f"is-dry-run: {dry_run}\n"
    print(result)
