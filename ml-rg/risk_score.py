#!/usr/bin/env python3
"""
PoC — Detecção de risco em jogo responsável (RG).

Heurística + score simples, 100% stdlib, sem dependências externas.
Em produção substitua por GradientBoosting/IsolationForest sobre features
do event-bus (apostas, depósitos, sessões) com treino em dados rotulados.

Escala retornada:
    0-30  Verde  — monitoramento normal
    30-60 Amarelo— alerta leve
    60+   Vermelho — intervenção obrigatória
"""

from __future__ import annotations

import argparse
import json
import sys
from dataclasses import dataclass, field
from datetime import datetime, time as dtime, timedelta

VERDE_LIM = 30
AMARELO_LIM = 60


@dataclass
class Session:
    started_at: datetime
    bets: list[float] = field(default_factory=list)
    deposits: list[float] = field(default_factory=list)
    losses: list[float] = field(default_factory=list)

    def hours(self, now: datetime) -> float:
        return (now - self.started_at).total_seconds() / 3600

    @property
    def persists_losses(self) -> bool:
        """Sinais de 'chasing': stakes crescentes após perdas consecutivas."""
        if len(self.losses) < 2:
            return False
        return self.losses[-1] > self.losses[-2] > 0


def _late_night_deposit_fraction(sessions: list[Session]) -> float:
    """% de depósitos feitos 00h-06h."""
    if not any(s.deposits for s in sessions):
        return 0.0
    total = sum(len(s.deposits) for s in sessions)
    late = sum(
        1
        for s in sessions
        for d in s.deposits
        if dtime(0, 0) <= s.started_at.time() <= dtime(5, 59)
    )
    return late / total


def compute_score(sessions: list[Session], now: datetime | None = None) -> tuple[int, list[str]]:
    flags: list[str] = []
    score = 0
    now = now or datetime.now()

    # Financeiro
    total_loss = sum(sum(s.losses) for s in sessions)
    stakes = [b for s in sessions for b in s.bets]
    total_stake = sum(stakes)

    if total_loss > 0 and total_stake > 0 and total_loss / total_stake > 0.6:
        score += 25
        flags.append("PERDA>60% DO VOLUME")
    if total_loss > 2000:
        score += 15
        flags.append("PERDA MENSAL RELEVANTE (>R$2k)")

    late = _late_night_deposit_fraction(sessions)
    if late > 0.3:
        score += 10
        flags.append("DEPOSITOS 00h-06h FREQUENTES")

    # Comportamental
    longest = max((s.hours(now) for s in sessions), default=0)
    if longest > 4:
        score += 15
        flags.append(f"SESSAO LONGA ({longest:.1f}h)")
    if any(s.persists_losses for s in sessions):
        score += 20
        flags.append("PERSECUCAO DE PERDAS (STAKE CRESCENTE)")
    if len(stakes) > 20:
        score += 10
        flags.append("ALTA FREQUENCIA DE APOSTAS (>20)")

    if total_stake > 5000:
        score += 10
        flags.append("VOLUME ALTO (>R$5k)")

    return min(score, 100), flags


def classify(score: int) -> str:
    if score >= AMARELO_LIM:
        return "VERMELHO"
    if score >= VERDE_LIM:
        return "AMARELO"
    return "VERDE"


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument(
        "sessions",
        nargs="?",
        type=argparse.FileType("r"),
        help="JSON file: [{\"started_at\": \"ISO\", \"bets\": [..], "
        "\"deposits\": [..], \"losses\": [..]}, ...]",
    )
    args = p.parse_args()

    raw = json.load(args or sys.stdin) if args.sessions else _demo_data()
    sessions = []
    for s in raw:
        sessions.append(
            Session(
                started_at=datetime.fromisoformat(s.get("started_at", "2026-09-01T00:00:00")),
                bets=s.get("bets", []),
                deposits=s.get("deposits", []),
                losses=s.get("losses", []),
            )
        )

    now = datetime.now()
    if args.sessions is None:
        # demo data is anchored "ago/hoje"
        now = max(s.started_at for s in sessions).replace(hour=23, minute=59)

    score, flags = compute_score(sessions, now=now)
    print(f"risk_score={score}  nivel={classify(score)}")
    for f in flags:
        print(f"  flag: {f}")
    return 0


def _demo_data() -> list[dict]:
    now = datetime.now()
    return [
        {
            "started_at": (now.replace(hour=2, minute=30, second=0) - timedelta(days=5)).isoformat(),
            "bets": [10, 20, 40, 80],
            "deposits": [100, 200],
            "losses": [10, 20, 40, 80],
        },
        {
            "started_at": (now.replace(hour=3, minute=15, second=0) - timedelta(days=3)).isoformat(),
            "bets": [20, 50, 120],
            "deposits": [150],
            "losses": [20, 50, 120],
        },
        {
            "started_at": (now.replace(hour=14, minute=0, second=0) - timedelta(days=1)).isoformat(),
            "bets": [5, 10, 15],
            "deposits": [50],
            "losses": [2, 4, 6],
        },
    ]


if __name__ == "__main__":
    sys.exit(main())