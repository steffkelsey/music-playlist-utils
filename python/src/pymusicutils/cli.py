import typer

from .pymsctl import pymsctl


app = typer.Typer()
app.command()(pymsctl)


if __name__ == "__main__":
    app()
