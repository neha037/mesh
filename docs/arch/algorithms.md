# Mesh: Key Algorithms

## FSRS Spaced Repetition (Pseudocode)

```
// FSRS v5 parameters
INITIAL_STABILITY = [0.4, 0.6, 2.4, 5.8]

function calculateRetrievability(stability, elapsedDays):
    return (1 + FACTOR * elapsedDays / stability) ^ DECAY

function scheduleReview(card, rating):
    elapsed = daysSince(card.lastReview)
    ret = calculateRetrievability(card.stability, elapsed)
    card.stability = calculateNextStability(card.difficulty, card.stability, ret, rating)
    card.difficulty = calculateNextDifficulty(card.difficulty, rating)
    card.dueDate = now() + card.stability days
    return card
```

## Kotkov Serendipity Metric

```
Serendipity(recommendations, history) =
    |{item ∈ recs : isRelevant(item) AND isDissimilar(item, history)}| / |recs|

Target: 0.3 - 0.5 (30-50% serendipitous suggestions)
```

## Cluster Density Scoring

```
For each tag cluster C:
    density(C) = |edges in C| / |max possible edges|
    centroid(C) = mean(embeddings in C)
    isolation(C) = min(distance to other centroids)
```
